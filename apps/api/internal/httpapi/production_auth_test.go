package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/auth"
)

type productionTokenVerifier struct {
	principal auth.Principal
	err       error
}

func (f productionTokenVerifier) Verify(context.Context, string) (auth.Principal, error) {
	return f.principal, f.err
}

type productionIdentityRepository struct {
	userID string
	err    error
}

func (f productionIdentityRepository) ResolveIdentity(context.Context, string, string, string) (string, error) {
	return f.userID, f.err
}

func TestProductionBearerAuthenticationPropagatesInternalUser(t *testing.T) {
	authentication := auth.NewAuthenticator(
		productionTokenVerifier{principal: auth.Principal{Issuer: "https://issuer.example", Subject: "subject-1"}},
		productionIdentityRepository{userID: "user-1"},
	)
	authorization := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}})
	handler := requirePermissionWithAuthentication(false, authentication, authorization, auth.ManagePortfolio, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userIDFromContext(r.Context())
		if !ok || userID != "user-1" {
			t.Fatalf("verified internal user was not propagated")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer signed-token")
	request.Header.Set("X-Organization-ID", "org-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected request to pass, got %d: %s", response.Code, response.Body.String())
	}
}

func TestProductionBearerAuthenticationRejectsInvalidToken(t *testing.T) {
	authentication := auth.NewAuthenticator(
		productionTokenVerifier{err: auth.ErrInvalidToken},
		productionIdentityRepository{},
	)
	authorization := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}})
	handler := requirePermissionWithAuthentication(false, authentication, authorization, auth.ViewPortfolio, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer bad-token")
	request.Header.Set("X-Organization-ID", "org-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", response.Code)
	}
	if response.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatal("expected Bearer challenge")
	}
}

func TestProductionBearerAuthenticationRejectsUnlinkedIdentity(t *testing.T) {
	authentication := auth.NewAuthenticator(
		productionTokenVerifier{principal: auth.Principal{Issuer: "https://issuer.example", Subject: "subject-1"}},
		productionIdentityRepository{err: auth.ErrIdentityNotLinked},
	)
	authorization := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}})
	handler := requirePermissionWithAuthentication(false, authentication, authorization, auth.ViewPortfolio, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer signed-token")
	request.Header.Set("X-Organization-ID", "org-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", response.Code)
	}
}

func TestProductionRequiresBearerToken(t *testing.T) {
	authorization := auth.NewService(fakeAuthRepository{membership: auth.Membership{Role: "admin"}})
	handler := requirePermissionWithAuthentication(false, nil, authorization, auth.ViewPortfolio, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Organization-ID", "org-1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", response.Code)
	}
}

func TestBearerTokenParsing(t *testing.T) {
	if token, err := bearerToken("bearer abc.def.ghi"); err != nil || token != "abc.def.ghi" {
		t.Fatalf("expected bearer token, got %q err=%v", token, err)
	}
	if _, err := bearerToken("Basic abc"); err == nil {
		t.Fatal("expected non-Bearer scheme to fail")
	}
	if !errors.Is(auth.ErrInvalidToken, auth.ErrInvalidToken) {
		t.Fatal("sentinel sanity check")
	}
}
