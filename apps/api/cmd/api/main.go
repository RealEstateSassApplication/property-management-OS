package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/config"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/database"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/httpapi"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/leases"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/owners"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/properties"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/rent"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenancies"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenants"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/units"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()

	pool, err := database.Connect(startupCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	authorizationService := auth.NewService(auth.NewPostgresRepository(pool))
	propertyService := properties.NewService(properties.NewPostgresRepository(pool))
	unitService := units.NewService(units.NewPostgresRepository(pool))
	tenantService := tenants.NewService(tenants.NewPostgresRepository(pool))
	tenancyService := tenancies.NewService(tenancies.NewPostgresRepository(pool))
	leaseService := leases.NewService(leases.NewPostgresRepository(pool))
	ownerService := owners.NewService(owners.NewPostgresRepository(pool))
	rentService := rent.NewService(rent.NewPostgresRepository(pool))
	maintenanceService := maintenance.NewService(maintenance.NewPostgresRepository(pool))

	handler := httpapi.NewRouter(httpapi.Dependencies{
		Authorization:            authorizationService,
		Properties:               propertyService,
		Units:                    unitService,
		Tenants:                  tenantService,
		Tenancies:                tenancyService,
		Leases:                   leaseService,
		Owners:                   ownerService,
		Rent:                     rentService,
		Maintenance:              maintenanceService,
		AllowDevelopmentIdentity: cfg.Environment == "development" || cfg.Environment == "test",
	})

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("api server starting", "address", cfg.Address, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("api server stopped")
}
