package config

import "os"

type Config struct {
	HTTPAddr             string
	GRPCAuthEndpoint     string
	GRPCCustomerEndpoint string
	JWTPublicKey         string
}

func Load() *Config {
	return &Config{
		HTTPAddr:             getEnv("HTTP_ADDR", ":8080"),
		GRPCAuthEndpoint:     getEnv("GRPC_AUTH_ENDPOINT", "localhost:9090"),
		GRPCCustomerEndpoint: getEnv("GRPC_CUSTOMER_ENDPOINT", "localhost:9091"),
		JWTPublicKey:         getEnv("JWT_PUBLIC_KEY", "public.pem"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
