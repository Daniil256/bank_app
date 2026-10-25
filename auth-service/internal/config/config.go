package config

import "os"

type Config struct {
	GRPCAddr      string
	DBDSN         string
	KafkaBroker   string
	JWTPrivateKey string
}

func Load() *Config {
	return &Config{
		GRPCAddr:      getEnv("GRPC_ADDR", ":9090"),
		DBDSN:         getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"),
		KafkaBroker:   getEnv("KAFKA_BROKER", "localhost:9092"),
		JWTPrivateKey: getEnv("JWT_PRIVATE_KEY", "private.pem"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
