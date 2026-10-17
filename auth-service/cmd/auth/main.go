package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	authv2 "auth-service/internal/adapter/inbound/grpc"
	"auth-service/internal/adapter/outbound/kafka"
	"auth-service/internal/adapter/outbound/postgres"
	"auth-service/internal/app/usecase"
	"auth-service/internal/config"
	authv1 "auth-service/proto/auth/v1"

	_ "github.com/lib/pq"
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

	// Репозиторий, продюсер, use case
	repo := postgres.NewUserRepository(db)
	producer := kafka.NewProducer(cfg.KafkaBroker)
	registerUC := usecase.NewRegisterUseCase(repo, producer)
	loginUC := usecase.NewLoginUseCase(repo)

	// gRPC
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("listen:", err)
	}

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, authv2.NewAuthServer(registerUC, loginUC))

	// Health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("auth.v1.AuthService", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	go func() {
		log.Println("auth-service started on", cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("serve:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	grpcServer.GracefulStop()
	log.Println("stopped")
}
