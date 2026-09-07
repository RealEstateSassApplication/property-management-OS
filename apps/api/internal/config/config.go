package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment      string
	Address          string
	DatabaseURL      string
	OIDCIssuerURL    string
	OIDCAudience     string
	StorageBucket    string
	StorageRegion    string
	StorageEndpoint  string
	StoragePathStyle bool
}

func Load() Config {
	return Config{
		Environment: getEnv("APP_ENV", "development"), Address: getEnv("API_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"),
		OIDCIssuerURL: os.Getenv("OIDC_ISSUER_URL"), OIDCAudience: os.Getenv("OIDC_AUDIENCE"),
		StorageBucket: os.Getenv("STORAGE_BUCKET"), StorageRegion: getEnv("STORAGE_REGION", "us-east-1"), StorageEndpoint: os.Getenv("STORAGE_ENDPOINT"),
		StoragePathStyle: getBoolEnv("STORAGE_PATH_STYLE", false),
	}
}

func getEnv(key, fallback string) string { if value := os.Getenv(key); value != "" { return value }; return fallback }
func getBoolEnv(key string, fallback bool) bool { value := os.Getenv(key); if value == "" { return fallback }; parsed, err := strconv.ParseBool(value); if err != nil { return fallback }; return parsed }
