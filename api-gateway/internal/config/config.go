package config

import "os"

type Config struct {
	HTTPAddr     string
	GRPCEndpoint string
	JWTPublicKey string
}

func Load() *Config {
	return &Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		GRPCEndpoint: getEnv("GRPC_ENDPOINT", "localhost:9090"),
		JWTPublicKey: getEnv("JWT_PUBLIC_KEY", "public.pem"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
