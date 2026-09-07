package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
)

type fakeAuthRepository struct {
	membership auth.Membership
	err        error
}

func (f fakeAuthRepository) GetMembership(context.Context, string, string) (auth.Membership, error) {
	return f.membership, f.err
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	NewRouter(Dependencies{}).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json, got %q", got)
	}
	if !strings.Contains(res.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected healthy response, got %s", res.Body.String())
	}
}

func TestDevelopmentAuthorizationContext(t *testing.T) {
	authorization := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := organizationIDFromContext(r.Context())
		if !ok || organizationID != "org-1" {
			t.Fatalf("organization context was not propagated")
		}
		userID, ok := userIDFromContext(r.Context())
		if !ok || userID != "user-1" {
			t.Fatalf("user context was not propagated")
		}
		role, ok := roleFromContext(r.Context())
		if !ok || role != "admin" {
			t.Fatalf("role context was not propagated")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := requirePermission(true, authorization, auth.ManagePortfolio, next)

	withoutHeaders := httptest.NewRecorder()
	handler.ServeHTTP(withoutHeaders, httptest.NewRequest(http.MethodGet, "/", nil))
	if withoutHeaders.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized without identity headers, got %d", withoutHeaders.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Organization-ID", "org-1")
	request.Header.Set("X-User-ID", "user-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected request to pass middleware, got %d", response.Code)
	}
}

func TestDevelopmentAuthorizationRejectsViewerWrite(t *testing.T) {
	authorization := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "viewer"}})
	handler := requirePermission(true, authorization, auth.ManagePeople, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("X-Organization-ID", "org-1")
	request.Header.Set("X-User-ID", "user-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", response.Code)
	}
}

func TestDevelopmentIdentityDisabled(t *testing.T) {
	authorization := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}})
	handler := requirePermission(false, authorization, auth.ViewPortfolio, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Organization-ID", "org-1")
	request.Header.Set("X-User-ID", "user-1")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected development identity to be disabled, got %d", response.Code)
	}
}
