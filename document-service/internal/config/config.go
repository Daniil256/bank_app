package config

import "os"

type Config struct {
	GRPCAddr       string
	DBDSN          string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
}

func Load() *Config {
	return &Config{
		GRPCAddr:       getEnv("GRPC_ADDR", ":9095"),
		DBDSN:          getEnv("DB_DSN", "postgres://postgres:postgres@127.0.0.1:5435/document_db?sslmode=disable"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "documents"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
