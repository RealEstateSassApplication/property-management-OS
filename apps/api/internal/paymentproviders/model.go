package paymentproviders

import (
	"errors"
	"time"
)

type PaidEvent struct {
	EventID        string `json:"eventId"`
	EventType      string `json:"eventType"`
	OrganizationID string `json:"organizationId"`
	TenantID       string `json:"tenantId"`
	AmountMinor    int64  `json:"amountMinor"`
	Currency       string `json:"currency"`
	ReceivedAt     string `json:"receivedAt"`
	ReferenceCode  string `json:"referenceCode"`
}

type ProcessResult struct {
	Provider  string    `json:"provider"`
	EventID   string    `json:"eventId"`
	Status    string    `json:"status"`
	PaymentID string    `json:"paymentId,omitempty"`
	Duplicate bool      `json:"duplicate"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	ErrInvalidSignature     = errors.New("invalid webhook signature")
	ErrInvalidEvent         = errors.New("invalid payment provider event")
	ErrEventPayloadConflict = errors.New("payment provider event id was reused with different payload")
	ErrTenantNotFound       = errors.New("tenant not found")
	ErrDuplicateRef         = errors.New("payment reference already exists")
)
