package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/paymentproviders"
)

type paymentWebhookRepository struct {
	result paymentproviders.ProcessResult
	err    error
}

func (r *paymentWebhookRepository) ProcessPaidEvent(context.Context, string, string, paymentproviders.PaidEvent) (paymentproviders.ProcessResult, error) {
	return r.result, r.err
}

func webhookSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func webhookBody() []byte {
	return []byte(`{"eventId":"evt-http-1","eventType":"payment.paid","organizationId":"11111111-1111-1111-1111-111111111111","tenantId":"66666666-6666-6666-6666-666666666666","amountMinor":15000000,"currency":"LKR","receivedAt":"2026-09-07","referenceCode":"HTTP-GW-001"}`)
}

func TestPaymentWebhookAcceptsValidSignedEvent(t *testing.T) {
	service := paymentproviders.NewService(&paymentWebhookRepository{result: paymentproviders.ProcessResult{Provider: "generic_hmac", EventID: "evt-http-1", Status: "processed", PaymentID: "payment-1"}}, "secret")
	handler := WithProductionRoutes(http.NotFoundHandler(), nil, service)
	body := webhookBody()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/payments/generic-hmac/webhook", strings.NewReader(string(body)))
	req.Header.Set("X-PropertyOS-Signature", webhookSignature("secret", body))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if res.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("expected security headers on webhook response")
	}
}

func TestPaymentWebhookRejectsInvalidSignature(t *testing.T) {
	service := paymentproviders.NewService(&paymentWebhookRepository{}, "secret")
	handler := WithProductionRoutes(http.NotFoundHandler(), nil, service)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/payments/generic-hmac/webhook", strings.NewReader(string(webhookBody())))
	req.Header.Set("X-PropertyOS-Signature", "sha256=00")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestPaymentWebhookReturnsUnavailableWhenDisabled(t *testing.T) {
	service := paymentproviders.NewService(&paymentWebhookRepository{}, "")
	handler := WithProductionRoutes(http.NotFoundHandler(), nil, service)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/payments/generic-hmac/webhook", strings.NewReader("{}"))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.Code)
	}
}

func TestPaymentWebhookRejectsOversizedBody(t *testing.T) {
	service := paymentproviders.NewService(&paymentWebhookRepository{}, "secret")
	handler := WithProductionRoutes(http.NotFoundHandler(), nil, service)
	body := strings.Repeat("x", maxWebhookBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/payments/generic-hmac/webhook", strings.NewReader(body))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", res.Code)
	}
}

func TestReadinessFailsWithoutDatabase(t *testing.T) {
	handler := WithProductionRoutes(http.NotFoundHandler(), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.Code)
	}
}
