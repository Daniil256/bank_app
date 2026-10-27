package config

import "os"

type Config struct {
	GRPCAddr     string
	DBDSN        string
	KafkaBrokers string
}

func Load() *Config {
	return &Config{
		GRPCAddr:     getEnv("GRPC_ADDR", ":9091"),
		DBDSN:        getEnv("DB_DSN", "postgres://postgres:postgres@127.0.0.1:5433/customer_db?sslmode=disable"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
