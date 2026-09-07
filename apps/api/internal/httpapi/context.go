package httpapi

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const organizationIDKey contextKey = "organization-id"

func organizationIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(organizationIDKey).(string)
	return value, ok && value != ""
}

func requireOrganizationContext(allowDevelopmentIdentity bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowDevelopmentIdentity {
			writeError(w, http.StatusServiceUnavailable, "authentication_not_configured", "authenticated organization context is not configured for this environment")
			return
		}

		organizationID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))
		if organizationID == "" {
			writeError(w, http.StatusUnauthorized, "organization_required", "X-Organization-ID is required in development")
			return
		}

		ctx := context.WithValue(r.Context(), organizationIDKey, organizationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
