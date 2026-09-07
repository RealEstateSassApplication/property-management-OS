package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
)

type contextKey string

const (
	organizationIDKey contextKey = "organization-id"
	userIDKey         contextKey = "user-id"
	roleKey           contextKey = "role"
)

func organizationIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(organizationIDKey).(string)
	return value, ok && value != ""
}

func userIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(userIDKey).(string)
	return value, ok && value != ""
}

func roleFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(roleKey).(string)
	return value, ok && value != ""
}

func requirePermission(allowDevelopmentIdentity bool, authorization *auth.Service, permission auth.Permission, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowDevelopmentIdentity {
			writeError(w, http.StatusServiceUnavailable, "authentication_not_configured", "production identity provider is not configured for this environment")
			return
		}
		if authorization == nil {
			writeError(w, http.StatusServiceUnavailable, "authorization_not_configured", "authorization service is not configured")
			return
		}

		organizationID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))
		userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
		if organizationID == "" || userID == "" {
			writeError(w, http.StatusUnauthorized, "identity_required", "X-Organization-ID and X-User-ID are required in development")
			return
		}

		membership, err := authorization.Authorize(r.Context(), organizationID, userID, permission)
		if errors.Is(err, auth.ErrMembershipNotFound) || errors.Is(err, auth.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "you do not have access to this organization resource")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "authorization_failed", "could not authorize request")
			return
		}

		ctx := context.WithValue(r.Context(), organizationIDKey, organizationID)
		ctx = context.WithValue(ctx, userIDKey, userID)
		ctx = context.WithValue(ctx, roleKey, membership.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
