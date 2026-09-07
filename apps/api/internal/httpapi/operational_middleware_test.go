package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeReadiness struct{ err error }

func (f fakeReadiness) Ping(context.Context) error { return f.err }

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestOperationalHandlerReadyWhenDatabaseResponds(t *testing.T) {
	handler := NewOperationalHandler(http.NotFoundHandler(), fakeReadiness{}, quietLogger(), OperationalConfig{})
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 readiness, got %d", response.Code)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected response request ID")
	}
}

func TestOperationalHandlerNotReadyWhenDatabaseFails(t *testing.T) {
	handler := NewOperationalHandler(http.NotFoundHandler(), fakeReadiness{err: errors.New("database unavailable")}, quietLogger(), OperationalConfig{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 readiness, got %d", response.Code)
	}
}

func TestOperationalHandlerPreservesValidRequestIDAndSecurityHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"ok": "true"}) })
	handler := NewOperationalHandler(next, fakeReadiness{}, quietLogger(), OperationalConfig{EnableHSTS: true})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", "trace-123")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Header().Get("X-Request-ID") != "trace-123" {
		t.Fatalf("expected request ID to be preserved, got %q", response.Header().Get("X-Request-ID"))
	}
	for _, header := range []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Strict-Transport-Security"} {
		if response.Header().Get(header) == "" {
			t.Fatalf("expected security header %s", header)
		}
	}
}

func TestOperationalHandlerRecoversPanics(t *testing.T) {
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })
	handler := NewOperationalHandler(next, fakeReadiness{}, quietLogger(), OperationalConfig{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/test", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after panic, got %d", response.Code)
	}
}

func TestOperationalHandlerLimitsMutationBody(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": "unexpected"})
			return
		}
		writeError(w, http.StatusRequestEntityTooLarge, "body_too_large", "request body exceeds limit")
	})
	handler := NewOperationalHandler(next, fakeReadiness{}, quietLogger(), OperationalConfig{MaxBodyBytes: 8})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader("0123456789")))
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized body, got %d", response.Code)
	}
}
