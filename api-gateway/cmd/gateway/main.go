package main

import (
	myHttp "api-gateway/internal/adapter/inbound/http"
	"api-gateway/internal/config"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/adapter/outbound/auth"
)

func main() {
	cfg := config.Load()

	// JWT-валидатор
	jwtValidator, err := auth.NewJWTValidator(cfg.JWTPublicKey)
	if err != nil {
		log.Fatal("jwt validator: ", err)
	}

	// gRPC-gateway mux
	ctx := context.Background()
	mux, err := myHttp.NewMux(ctx, cfg.GRPCEndpoint)
	if err != nil {
		log.Fatal("http mux: ", err)
	}

	// Оборачиваем в middleware
	handler := jwtValidator.Middleware(mux)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,
	}

	go func() {
		log.Println("api-gateway started on", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("shutdown error:", err)
	}
	log.Println("api-gateway stopped")
}
