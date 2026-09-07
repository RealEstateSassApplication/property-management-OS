package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/accounting"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/agentactions"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/config"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/database"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/documents"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/httpapi"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/inspections"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/leases"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/notifications"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/orgadmin"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/owners"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/paymentproviders"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/portals"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/properties"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/rent"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/reporting"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenancies"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenants"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/units"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := database.Connect(startupCtx, cfg.DatabaseURL)
	cancelStartup()
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	authRepository := auth.NewPostgresRepository(pool)
	authorizationService := auth.NewService(authRepository)
	allowDevelopmentIdentity := cfg.Environment == "development" || cfg.Environment == "test"
	var authenticationService *auth.Authenticator
	issuerConfigured := strings.TrimSpace(cfg.OIDCIssuerURL) != ""
	audienceConfigured := strings.TrimSpace(cfg.OIDCAudience) != ""
	if issuerConfigured != audienceConfigured {
		logger.Error("OIDC configuration is incomplete", "required", "OIDC_ISSUER_URL and OIDC_AUDIENCE must be configured together")
		os.Exit(1)
	}
	if issuerConfigured {
		authCtx, cancelAuth := context.WithTimeout(context.Background(), 10*time.Second)
		verifier, err := auth.NewOIDCVerifier(authCtx, cfg.OIDCIssuerURL, cfg.OIDCAudience)
		cancelAuth()
		if err != nil {
			logger.Error("OIDC provider initialization failed", "error", err)
			os.Exit(1)
		}
		authenticationService = auth.NewAuthenticator(verifier, authRepository)
	}
	if !allowDevelopmentIdentity && authenticationService == nil {
		logger.Error("production authentication is not configured", "required", "OIDC_ISSUER_URL and OIDC_AUDIENCE")
		os.Exit(1)
	}

	var documentStorage documents.Storage
	if strings.TrimSpace(cfg.StorageBucket) != "" {
		storageCtx, cancelStorage := context.WithTimeout(context.Background(), 10*time.Second)
		storage, err := documents.NewS3Storage(storageCtx, documents.S3StorageConfig{Bucket: cfg.StorageBucket, Region: cfg.StorageRegion, BaseEndpoint: cfg.StorageEndpoint, UsePathStyle: cfg.StoragePathStyle})
		cancelStorage()
		if err != nil {
			logger.Error("document storage initialization failed", "error", err)
			os.Exit(1)
		}
		documentStorage = storage
	} else if !allowDevelopmentIdentity {
		logger.Error("production document storage is not configured", "required", "STORAGE_BUCKET")
		os.Exit(1)
	} else {
		logger.Warn("document storage disabled; document upload/download endpoints will return 503")
	}

	propertyService := properties.NewService(properties.NewPostgresRepository(pool))
	unitService := units.NewService(units.NewPostgresRepository(pool))
	tenantService := tenants.NewService(tenants.NewPostgresRepository(pool))
	tenancyService := tenancies.NewService(tenancies.NewPostgresRepository(pool))
	leaseService := leases.NewService(leases.NewPostgresRepository(pool))
	ownerService := owners.NewService(owners.NewPostgresRepository(pool))
	rentService := rent.NewService(rent.NewPostgresRepository(pool))
	accountingService := accounting.NewService(accounting.NewPostgresRepository(pool))
	inspectionService := inspections.NewService(inspections.NewPostgresRepository(pool))
	maintenanceService := maintenance.NewService(maintenance.NewPostgresRepository(pool))
	documentService := documents.NewService(documents.NewPostgresRepository(pool), documentStorage)
	notificationService := notifications.NewService(notifications.NewPostgresRepository(pool))
	agentActionService := agentactions.NewService(agentactions.NewPostgresRepository(pool), agentactions.NewMaintenanceExecutor(maintenanceService))
	portalService := portals.NewService(portals.NewPostgresRepository(pool), maintenanceService)
	organizationAdminService := orgadmin.NewService(orgadmin.NewPostgresRepository(pool))
	reportingService := reporting.NewService(reporting.NewPostgresRepository(pool))
	paymentProviderService := paymentproviders.NewService(paymentproviders.NewPostgresRepository(pool), cfg.PaymentWebhookSecret)

	handler := httpapi.NewRouter(httpapi.Dependencies{
		Authentication: authenticationService, Authorization: authorizationService,
		Properties: propertyService, Units: unitService, Tenants: tenantService,
		Tenancies: tenancyService, Leases: leaseService, Owners: ownerService,
		Rent: rentService, Accounting: accountingService, Maintenance: maintenanceService,
		Documents: documentService, Notifications: notificationService, AgentActions: agentActionService,
		Portals: portalService, OrganizationAdmin: organizationAdminService, Reporting: reportingService,
		AllowDevelopmentIdentity: allowDevelopmentIdentity,
	})
	handler = httpapi.WithInspectionRoutes(handler, authenticationService, authorizationService, allowDevelopmentIdentity, inspectionService)
	handler = httpapi.WithProductionRoutes(handler, pool, paymentProviderService)

	server := &http.Server{Addr: cfg.Address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("api server starting", "address", cfg.Address, "environment", cfg.Environment, "oidc", authenticationService != nil, "documentStorage", documentStorage != nil, "paymentWebhook", paymentProviderService.Enabled())
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
