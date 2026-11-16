package config

import "os"

type Config struct {
	GRPCAddr         string
	DBDSN            string
	KafkaBroker      string
	CustomerEndpoint string
	DocumentEndpoint string
}

func Load() *Config {
	return &Config{
		GRPCAddr:         getEnv("GRPC_ADDR", ":9094"),
		DBDSN:            getEnv("DB_DSN", "postgres://postgres:postgres@127.0.0.1:5434/kyc_db?sslmode=disable"),
		KafkaBroker:      getEnv("KAFKA_BROKER", "localhost:9092"),
		CustomerEndpoint: getEnv("CUSTOMER_ENDPOINT", "localhost:9091"),
		DocumentEndpoint: getEnv("DOCUMENT_ENDPOINT", "localhost:9093"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
