package config

import "os"

type Config struct {
	Environment string
	Address     string
	DatabaseURL string
}

func Load() Config {
	return Config{
		Environment: getEnv("APP_ENV", "development"),
		Address:     getEnv("API_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
