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
	return requirePermissionWithAuthentication(allowDevelopmentIdentity, nil, authorization, permission, next)
}

func requirePermissionWithAuthentication(allowDevelopmentIdentity bool, authentication *auth.Authenticator, authorization *auth.Service, permission auth.Permission, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authorization == nil {
			writeError(w, http.StatusServiceUnavailable, "authorization_not_configured", "authorization service is not configured")
			return
		}

		organizationID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))
		if organizationID == "" {
			writeError(w, http.StatusBadRequest, "organization_required", "X-Organization-ID is required")
			return
		}

		userID, ok := authenticateRequest(w, r, allowDevelopmentIdentity, authentication)
		if !ok {
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

func authenticateRequest(w http.ResponseWriter, r *http.Request, allowDevelopmentIdentity bool, authentication *auth.Authenticator) (string, bool) {
	authorizationHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorizationHeader != "" {
		rawToken, err := bearerToken(authorizationHeader)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "invalid_authorization", err.Error())
			return "", false
		}
		if authentication == nil {
			writeError(w, http.StatusServiceUnavailable, "authentication_not_configured", "OIDC authentication is not configured")
			return "", false
		}
		identity, err := authentication.Authenticate(r.Context(), rawToken)
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "invalid_token", "bearer token could not be verified")
			return "", false
		case errors.Is(err, auth.ErrIdentityNotLinked):
			writeError(w, http.StatusForbidden, "identity_not_linked", "verified identity is not linked to a Property OS user")
			return "", false
		case errors.Is(err, auth.ErrAuthenticationNotConfigured):
			writeError(w, http.StatusServiceUnavailable, "authentication_not_configured", "OIDC authentication is not configured")
			return "", false
		case err != nil:
			writeError(w, http.StatusInternalServerError, "authentication_failed", "could not authenticate request")
			return "", false
		}
		return identity.UserID, true
	}

	if allowDevelopmentIdentity {
		userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "identity_required", "X-User-ID is required in development when no Bearer token is supplied")
			return "", false
		}
		return userID, true
	}

	w.Header().Set("WWW-Authenticate", "Bearer")
	writeError(w, http.StatusUnauthorized, "bearer_token_required", "Authorization: Bearer <token> is required")
	return "", false
}

func bearerToken(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("Authorization header must use the Bearer scheme")
	}
	return parts[1], nil
}
