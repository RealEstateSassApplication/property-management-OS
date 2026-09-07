package notifications

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	List(ctx context.Context, organizationID string) ([]Notification, error)
	Enqueue(ctx context.Context, organizationID, actorUserID string, input EnqueueInput) (Notification, error)
	GetRentReminderContext(ctx context.Context, organizationID, obligationID string) (RentReminderContext, error)
	ClaimBatch(ctx context.Context, workerID string, limit int) ([]Notification, error)
	MarkDelivered(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, failure string, retryAt time.Time, dead bool) error
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, organizationID string) ([]Notification, error) {
	return s.repository.List(ctx, organizationID)
}

func (s *Service) Enqueue(ctx context.Context, organizationID, actorUserID string, input EnqueueInput) (Notification, error) {
	input.Topic = strings.TrimSpace(input.Topic)
	input.Channel = strings.ToLower(strings.TrimSpace(input.Channel))
	input.Recipient = strings.TrimSpace(input.Recipient)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Body = strings.TrimSpace(input.Body)
	input.ResourceType = strings.TrimSpace(input.ResourceType)
	input.ResourceID = strings.TrimSpace(input.ResourceID)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.Topic == "" {
		return Notification{}, ErrTopicRequired
	}
	if !validChannel(input.Channel) {
		return Notification{}, ErrInvalidChannel
	}
	if input.Recipient == "" {
		return Notification{}, ErrRecipientRequired
	}
	if input.Body == "" {
		return Notification{}, ErrBodyRequired
	}
	if (input.ResourceType == "") != (input.ResourceID == "") {
		return Notification{}, ErrResourcePairRequired
	}
	return s.repository.Enqueue(ctx, organizationID, actorUserID, input)
}

func (s *Service) QueueRentReminder(ctx context.Context, organizationID, actorUserID string, input RentReminderInput) (Notification, error) {
	input.ObligationID = strings.TrimSpace(input.ObligationID)
	input.Channel = strings.ToLower(strings.TrimSpace(input.Channel))
	input.Recipient = strings.TrimSpace(input.Recipient)
	if input.ObligationID == "" {
		return Notification{}, ErrObligationNotFound
	}
	if input.Channel != "email" && input.Channel != "sms" && input.Channel != "whatsapp" {
		return Notification{}, ErrInvalidChannel
	}
	if input.Recipient == "" {
		return Notification{}, ErrRecipientRequired
	}
	context, err := s.repository.GetRentReminderContext(ctx, organizationID, input.ObligationID)
	if err != nil {
		return Notification{}, err
	}
	if context.BalanceMinor <= 0 || context.State == "paid" || context.State == "void" {
		return Notification{}, ErrRentReminderNotApplicable
	}
	amount := fmt.Sprintf("%s %.2f", context.Currency, float64(context.BalanceMinor)/100)
	subject := fmt.Sprintf("Rent reminder: %s outstanding", amount)
	body := fmt.Sprintf("Hello %s, %s remains outstanding for %s · %s for %s and was due on %s. Please arrange payment or contact the property manager if payment has already been made.", context.TenantName, amount, context.PropertyName, context.UnitLabel, context.Period, context.DueDate)
	key := fmt.Sprintf("rent-reminder:%s:%s:%s:%s", context.ObligationID, input.Channel, strings.ToLower(input.Recipient), time.Now().UTC().Format("2006-01-02"))
	return s.Enqueue(ctx, organizationID, actorUserID, EnqueueInput{
		Topic:          "rent.reminder",
		Channel:        input.Channel,
		Recipient:      input.Recipient,
		Subject:        subject,
		Body:           body,
		ResourceType:   "rent_obligation",
		ResourceID:     context.ObligationID,
		IdempotencyKey: key,
	})
}

func validChannel(channel string) bool {
	switch channel {
	case "email", "sms", "whatsapp", "webhook":
		return true
	default:
		return false
	}
}
