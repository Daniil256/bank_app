package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcadapter "customer-service/internal/adapter/inbound/grpc"
	"customer-service/internal/adapter/outbound/kafka"
	"customer-service/internal/adapter/outbound/postgres"
	"customer-service/internal/app/usecase"
	"customer-service/internal/config"
	"customer-service/migrations"
	customerv1 "customer-service/proto/customer/v1"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	// БД
	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}

	// Миграции
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("goose dialect:", err)
	}
	if err := goose.Up(db, "."); err != nil {
		log.Fatal("migrate:", err)
	}

	// Репозиторий
	repo := postgres.NewCustomerRepository(db)

	// Kafka
	producer := kafka.NewProducer(cfg.KafkaBrokers)

	// Use cases
	listUC := usecase.NewListCustomersUseCase(repo)
	getCustomerUC := usecase.NewGetCustomerUseCase(repo)
	updateProfileUC := usecase.NewUpdateProfileUseCase(repo, producer)
	getStatusUC := usecase.NewGetStatusUseCase(repo)

	// gRPC
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("listen:", err)
	}
	grpcServer := grpc.NewServer()
	customerv1.RegisterCustomerServiceServer(grpcServer,
		grpcadapter.NewCustomerServer(getCustomerUC, updateProfileUC, getStatusUC, listUC))

	// Kafka consumer (заглушка)
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, repo)
	go func() {
		if err := consumer.Run(context.Background()); err != nil {
			log.Printf("[kafka.consumer] error: %v", err)
		}
	}()

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("customer.v1.CustomerService", grpc_health_v1.HealthCheckResponse_SERVING)
	reflection.Register(grpcServer)

	go func() {
		log.Println("customer-service started on", cfg.GRPCAddr)
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
