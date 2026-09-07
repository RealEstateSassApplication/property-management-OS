package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/paymentproviders"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxWebhookBodyBytes = 1 << 20

func WithProductionRoutes(base http.Handler, pool *pgxpool.Pool, payments *paymentproviders.Service) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if pool == nil || pool.Ping(ctx) != nil {
			writeError(w, http.StatusServiceUnavailable, "database_unavailable", "database readiness check failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "service": "property-management-os-api"})
	})

	mux.HandleFunc("POST /api/v1/integrations/payments/generic-hmac/webhook", func(w http.ResponseWriter, r *http.Request) {
		if payments == nil || !payments.Enabled() {
			writeError(w, http.StatusServiceUnavailable, "payment_provider_disabled", "payment provider webhook is not configured")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusRequestEntityTooLarge, "webhook_body_invalid", "webhook body exceeds the allowed size")
			return
		}
		result, err := payments.VerifyAndProcess(r.Context(), "generic_hmac", r.Header.Get("X-PropertyOS-Signature"), body)
		switch {
		case errors.Is(err, paymentproviders.ErrInvalidSignature):
			writeError(w, http.StatusUnauthorized, "webhook_signature_invalid", err.Error())
			return
		case errors.Is(err, paymentproviders.ErrInvalidEvent):
			writeError(w, http.StatusBadRequest, "payment_event_invalid", err.Error())
			return
		case errors.Is(err, paymentproviders.ErrEventPayloadConflict):
			writeError(w, http.StatusConflict, "payment_event_id_conflict", err.Error())
			return
		case errors.Is(err, paymentproviders.ErrTenantNotFound), errors.Is(err, paymentproviders.ErrDuplicateRef):
			writeError(w, http.StatusConflict, "payment_event_rejected", err.Error())
			return
		case err != nil:
			writeError(w, http.StatusInternalServerError, "payment_event_failed", "payment provider event could not be processed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": result})
	})

	mux.Handle("/", base)
	return observabilityMiddleware(securityHeaders(mux))
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
