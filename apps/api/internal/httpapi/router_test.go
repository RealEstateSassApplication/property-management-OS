package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestDevelopmentOrganizationContext(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := organizationIDFromContext(r.Context())
		if !ok || organizationID != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("organization context was not propagated")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := requireOrganizationContext(true, next)

	withoutHeader := httptest.NewRecorder()
	handler.ServeHTTP(withoutHeader, httptest.NewRequest(http.MethodGet, "/", nil))
	if withoutHeader.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized without organization header, got %d", withoutHeader.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Organization-ID", "11111111-1111-1111-1111-111111111111")
	withHeader := httptest.NewRecorder()
	handler.ServeHTTP(withHeader, request)
	if withHeader.Code != http.StatusNoContent {
		t.Fatalf("expected request to pass middleware, got %d", withHeader.Code)
	}
}

func TestDevelopmentIdentityDisabled(t *testing.T) {
	handler := requireOrganizationContext(false, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Organization-ID", "11111111-1111-1111-1111-111111111111")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected development identity to be disabled, got %d", response.Code)
	}
}
