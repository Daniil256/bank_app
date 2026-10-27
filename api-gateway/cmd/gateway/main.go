package main

import (
	myHttp "api-gateway/internal/adapter/inbound/http"
	"api-gateway/internal/config"
	"context"
	"errors"
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
	mux, err := myHttp.NewMux(ctx, cfg.GRPCAuthEndpoint, cfg.GRPCCustomerEndpoint)
	if err != nil {
		log.Fatal("http mux: ", err)
	}

	// Оборачиваем в middleware
	handler := mux
	handler = jwtValidator.Middleware(handler) // JWT
	//handler = rateLimiter.Middleware(handler)        // Rate limit
	//handler = requestIDMiddleware(handler)           // Request-ID
	//handler = accessLogMiddleware(handler)           // Access-логи
	handler = auth.CORS(handler) // CORS — САМЫЙ ВНЕШНИЙ
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,
	}

	go func() {
		log.Println("api-gateway started on", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
