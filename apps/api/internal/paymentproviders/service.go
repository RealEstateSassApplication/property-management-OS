package paymentproviders

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

type Repository interface {
	ProcessPaidEvent(ctx context.Context, provider string, payloadHash string, event PaidEvent) (ProcessResult, error)
}

type Service struct {
	repository Repository
	secret     []byte
}

func NewService(repository Repository, secret string) *Service {
	return &Service{repository: repository, secret: []byte(strings.TrimSpace(secret))}
}

func (s *Service) Enabled() bool { return len(s.secret) > 0 }

func (s *Service) VerifyAndProcess(ctx context.Context, provider, signature string, body []byte) (ProcessResult, error) {
	if !s.Enabled() || !verifySignature(s.secret, signature, body) {
		return ProcessResult{}, ErrInvalidSignature
	}
	var event PaidEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return ProcessResult{}, ErrInvalidEvent
	}
	event.EventID = strings.TrimSpace(event.EventID)
	event.EventType = strings.ToLower(strings.TrimSpace(event.EventType))
	event.OrganizationID = strings.TrimSpace(event.OrganizationID)
	event.TenantID = strings.TrimSpace(event.TenantID)
	event.Currency = strings.ToUpper(strings.TrimSpace(event.Currency))
	event.ReceivedAt = strings.TrimSpace(event.ReceivedAt)
	event.ReferenceCode = strings.TrimSpace(event.ReferenceCode)
	if event.EventID == "" || event.EventType != "payment.paid" || event.OrganizationID == "" || event.TenantID == "" || event.AmountMinor <= 0 || len(event.Currency) != 3 || event.ReferenceCode == "" {
		return ProcessResult{}, ErrInvalidEvent
	}
	if _, err := time.Parse("2006-01-02", event.ReceivedAt); err != nil {
		return ProcessResult{}, ErrInvalidEvent
	}
	hash := sha256.Sum256(body)
	return s.repository.ProcessPaidEvent(ctx, provider, hex.EncodeToString(hash[:]), event)
}

func verifySignature(secret []byte, signature string, body []byte) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return false
	}
	provided, err := hex.DecodeString(strings.TrimPrefix(signature, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(body)
	return hmac.Equal(provided, mac.Sum(nil))
}
