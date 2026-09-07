package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/config"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/database"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/notifications"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	startupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	pool, err := database.Connect(startupCtx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		logger.Error("worker database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	providerName := strings.ToLower(strings.TrimSpace(os.Getenv("NOTIFICATION_PROVIDER")))
	if providerName == "" {
		providerName = "log"
	}
	var provider notifications.DeliveryProvider
	switch providerName {
	case "log":
		if cfg.Environment != "development" && cfg.Environment != "test" {
			logger.Error("log notification provider is development-only", "required", "NOTIFICATION_PROVIDER=webhook")
			os.Exit(1)
		}
		provider = notifications.NewLogProvider(logger)
	case "webhook":
		url := strings.TrimSpace(os.Getenv("NOTIFICATION_WEBHOOK_URL"))
		if url == "" {
			logger.Error("notification webhook URL is required")
			os.Exit(1)
		}
		provider = notifications.NewWebhookProvider(url, os.Getenv("NOTIFICATION_WEBHOOK_TOKEN"))
	default:
		logger.Error("unsupported notification provider", "provider", providerName)
		os.Exit(1)
	}

	poll := 5 * time.Second
	if raw := strings.TrimSpace(os.Getenv("NOTIFICATION_POLL_INTERVAL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed < time.Second {
			logger.Error("invalid NOTIFICATION_POLL_INTERVAL", "value", raw)
			os.Exit(1)
		}
		poll = parsed
	}
	host, _ := os.Hostname()
	workerID := fmt.Sprintf("%s:%d", host, os.Getpid())
	worker := notifications.NewWorker(notifications.NewPostgresRepository(pool), provider, logger, workerID)
	logger.Info("notification worker starting", "workerId", workerID, "provider", providerName, "pollInterval", poll)
	if err := worker.Run(ctx, poll); err != nil && ctx.Err() == nil {
		logger.Error("notification worker stopped unexpectedly", "error", err)
		os.Exit(1)
	}
	logger.Info("notification worker stopped")
}
