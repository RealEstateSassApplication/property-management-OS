package config

import "os"

type Config struct {
	Environment   string
	Address       string
	DatabaseURL   string
	OIDCIssuerURL string
	OIDCAudience  string
}

func Load() Config {
	return Config{
		Environment:   getEnv("APP_ENV", "development"),
		Address:       getEnv("API_ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		OIDCIssuerURL: os.Getenv("OIDC_ISSUER_URL"),
		OIDCAudience:  os.Getenv("OIDC_AUDIENCE"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
