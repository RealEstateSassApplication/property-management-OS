package notifications

import (
	"encoding/json"
	"errors"
	"time"
)

type Notification struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organizationId"`
	ActorUserID    string          `json:"actorUserId,omitempty"`
	Topic          string          `json:"topic"`
	Channel        string          `json:"channel"`
	Recipient      string          `json:"recipient"`
	Subject        string          `json:"subject,omitempty"`
	Body           string          `json:"body"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	ResourceType   string          `json:"resourceType,omitempty"`
	ResourceID     string          `json:"resourceId,omitempty"`
	IdempotencyKey string          `json:"idempotencyKey,omitempty"`
	Status         string          `json:"status"`
	AttemptCount   int             `json:"attemptCount"`
	MaxAttempts    int             `json:"maxAttempts"`
	AvailableAt    time.Time       `json:"availableAt"`
	LockedAt       *time.Time      `json:"lockedAt,omitempty"`
	LockedBy       string          `json:"lockedBy,omitempty"`
	LastError      string          `json:"lastError,omitempty"`
	DeliveredAt    *time.Time      `json:"deliveredAt,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

type EnqueueInput struct {
	Topic          string          `json:"topic"`
	Channel        string          `json:"channel"`
	Recipient      string          `json:"recipient"`
	Subject        string          `json:"subject"`
	Body           string          `json:"body"`
	Payload        json.RawMessage `json:"payload"`
	ResourceType   string          `json:"resourceType"`
	ResourceID     string          `json:"resourceId"`
	IdempotencyKey string          `json:"idempotencyKey"`
}

type RentReminderInput struct {
	ObligationID string `json:"obligationId"`
	Channel      string `json:"channel"`
	Recipient    string `json:"recipient"`
}

type RentReminderContext struct {
	ObligationID string
	TenantName   string
	PropertyName string
	UnitLabel    string
	Period       string
	DueDate      string
	BalanceMinor int64
	Currency     string
	State        string
}

var (
	ErrInvalidChannel            = errors.New("notification channel must be email, sms, whatsapp, or webhook")
	ErrRecipientRequired         = errors.New("notification recipient is required")
	ErrTopicRequired             = errors.New("notification topic is required")
	ErrBodyRequired              = errors.New("notification body is required")
	ErrResourcePairRequired      = errors.New("resourceType and resourceId must be supplied together")
	ErrObligationNotFound        = errors.New("rent obligation not found")
	ErrRentReminderNotApplicable = errors.New("rent reminder is not applicable to a paid or void obligation")
)
