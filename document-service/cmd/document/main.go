package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	mygrpc "document-service/internal/adapter/inbound/grpc"
	"document-service/internal/adapter/outbound/minio"
	"document-service/internal/adapter/outbound/postgres"
	"document-service/internal/app/usecase"
	"document-service/internal/config"
	"document-service/migrations"
	documentv1 "document-service/proto/document/v1"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db, "."); err != nil {
		log.Fatal("migrate:", err)
	}

	repo := postgres.NewDocumentRepository(db)

	storage, err := minio.NewClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, false)
	if err != nil {
		log.Fatal("minio:", err)
	}

	initUC := usecase.NewInitUploadUseCase(repo, storage)
	confirmUC := usecase.NewConfirmUploadUseCase(repo, storage)
	getURLUC := usecase.NewGetDownloadURLUseCase(repo, storage)
	listUC := usecase.NewListDocumentsUseCase(repo)
	hasRequiredUC := usecase.NewHasRequiredDocumentsUseCase(repo)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("listen:", err)
	}

	grpcServer := grpc.NewServer()
	documentv1.RegisterDocumentServiceServer(grpcServer,
		mygrpc.NewDocumentServer(initUC, confirmUC, getURLUC, listUC, hasRequiredUC),
	)
	reflection.Register(grpcServer)

	go func() {
		log.Println("document-service started on", cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("serve:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	grpcServer.GracefulStop()
}
