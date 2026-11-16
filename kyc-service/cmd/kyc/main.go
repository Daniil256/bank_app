package main

import (
	"context"
	"database/sql"
	"kyc-service/internal/app/worker"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	myGrpc "kyc-service/internal/adapter/inbound/grpc"
	"kyc-service/internal/adapter/outbound/customer"
	"kyc-service/internal/adapter/outbound/document"
	"kyc-service/internal/adapter/outbound/kafka"
	"kyc-service/internal/adapter/outbound/postgres"
	"kyc-service/internal/app/usecase"
	"kyc-service/internal/config"
	"kyc-service/migrations"
	kycv1 "kyc-service/proto/kyc/v1"

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

	repo := postgres.NewApplicationRepository(db)
	checkRepo := postgres.NewCheckRepository(db)
	sanctionsChecker := postgres.NewSanctionsChecker(db)

	customerClient, err := customer.NewClient(cfg.CustomerEndpoint)
	if err != nil {
		log.Fatal("customer client:", err)
	}
	defer customerClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	documentClient, err := document.NewClient(cfg.DocumentEndpoint)
	if err != nil {
		log.Fatal("document client:", err)
	}
	defer documentClient.Close()

	producer := kafka.NewProducer(cfg.KafkaBroker)

	submitUC := usecase.NewSubmitUseCase(repo, customerClient,
		documentClient,
		producer)
	getUC := usecase.NewGetUseCase(repo)
	getUCClient := usecase.NewGetClientUseCase(repo)

	listUC := usecase.NewListUseCase(repo)

	claimUC := usecase.NewClaimUseCase(repo)
	approveUC := usecase.NewApproveUseCase(repo, producer)
	rejectUC := usecase.NewRejectUseCase(repo, producer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("listen:", err)
	}

	worker := worker.NewAutoChecksWorker(repo, checkRepo, sanctionsChecker, customerClient)
	go worker.Run(ctx)

	grpcServer := grpc.NewServer()
	kycv1.RegisterKycServiceServer(grpcServer, myGrpc.NewKycServer(submitUC, getUC, listUC, getUCClient, claimUC, approveUC, rejectUC))
	reflection.Register(grpcServer)

	go func() {
		log.Println("kyc-service started on", cfg.GRPCAddr)
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
