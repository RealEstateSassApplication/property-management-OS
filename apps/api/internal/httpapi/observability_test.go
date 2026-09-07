package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestObservabilityMiddlewarePreservesSafeRequestID(t *testing.T) {
	var contextRequestID string
	handler := observabilityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextRequestID = requestIDFromContext(r.Context())
		w.WriteHeader(http.StatusAccepted)
	}))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(requestIDHeader, "edge-req_123")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", res.Code)
	}
	if res.Header().Get(requestIDHeader) != "edge-req_123" || contextRequestID != "edge-req_123" {
		t.Fatalf("request ID was not propagated: header=%q context=%q", res.Header().Get(requestIDHeader), contextRequestID)
	}
}

func TestObservabilityMiddlewareReplacesUnsafeRequestID(t *testing.T) {
	handler := observabilityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(requestIDHeader, "unsafe request id")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	requestID := res.Header().Get(requestIDHeader)
	if requestID == "" || requestID == "unsafe request id" {
		t.Fatalf("expected generated request ID, got %q", requestID)
	}
}
