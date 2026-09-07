package paymentproviders

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

type fakeRepository struct {
	provider    string
	payloadHash string
	event       PaidEvent
	result      ProcessResult
	err         error
}

func (f *fakeRepository) ProcessPaidEvent(_ context.Context, provider, payloadHash string, event PaidEvent) (ProcessResult, error) {
	f.provider = provider
	f.payloadHash = payloadHash
	f.event = event
	return f.result, f.err
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func validBody() []byte {
	return []byte(`{"eventId":"evt-1","eventType":"payment.paid","organizationId":"11111111-1111-1111-1111-111111111111","tenantId":"66666666-6666-6666-6666-666666666666","amountMinor":15000000,"currency":"lkr","receivedAt":"2026-09-07","referenceCode":"GW-001"}`)
}

func TestVerifyAndProcessAcceptsValidSignedEvent(t *testing.T) {
	repo := &fakeRepository{result: ProcessResult{Status: "processed", PaymentID: "payment-1"}}
	service := NewService(repo, "test-secret")
	body := validBody()

	result, err := service.VerifyAndProcess(context.Background(), "generic_hmac", sign("test-secret", body), body)
	if err != nil {
		t.Fatal(err)
	}
	if result.PaymentID != "payment-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.provider != "generic_hmac" || repo.event.Currency != "LKR" || repo.event.ReferenceCode != "GW-001" {
		t.Fatalf("unexpected normalized repository input: %+v", repo.event)
	}
	if len(repo.payloadHash) != 64 {
		t.Fatalf("expected SHA-256 payload hash, got %q", repo.payloadHash)
	}
}

func TestVerifyAndProcessRejectsInvalidSignature(t *testing.T) {
	service := NewService(&fakeRepository{}, "test-secret")
	_, err := service.VerifyAndProcess(context.Background(), "generic_hmac", "sha256=00", validBody())
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestVerifyAndProcessRejectsMalformedOrUnsupportedEvent(t *testing.T) {
	service := NewService(&fakeRepository{}, "test-secret")
	body := []byte(`{"eventId":"evt-1","eventType":"payment.failed"}`)
	_, err := service.VerifyAndProcess(context.Background(), "generic_hmac", sign("test-secret", body), body)
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestVerifyAndProcessRejectsInvalidDate(t *testing.T) {
	service := NewService(&fakeRepository{}, "test-secret")
	body := []byte(`{"eventId":"evt-1","eventType":"payment.paid","organizationId":"org","tenantId":"tenant","amountMinor":1,"currency":"LKR","receivedAt":"07-09-2026","referenceCode":"GW-001"}`)
	_, err := service.VerifyAndProcess(context.Background(), "generic_hmac", sign("test-secret", body), body)
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestDisabledServiceRejectsWebhook(t *testing.T) {
	service := NewService(&fakeRepository{}, "")
	if service.Enabled() {
		t.Fatal("expected service to be disabled")
	}
	_, err := service.VerifyAndProcess(context.Background(), "generic_hmac", "", validBody())
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}
