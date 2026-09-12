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

type PushDeviceRepository interface {
	RegisterPushDevice(ctx context.Context, organizationID, userID string, input RegisterPushDeviceInput) (PushDevice, error)
	ListPushDevices(ctx context.Context, organizationID, userID string) ([]PushDevice, error)
	DeletePushDevice(ctx context.Context, organizationID, userID, deviceID string) error
}

type Service struct {
	repository  Repository
	pushDevices PushDeviceRepository
}

func NewService(repository Repository) *Service {
	service := &Service{repository: repository}
	if pushDevices, ok := repository.(PushDeviceRepository); ok {
		service.pushDevices = pushDevices
	}
	return service
}

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

func (s *Service) RegisterPushDevice(ctx context.Context, organizationID, userID string, input RegisterPushDeviceInput) (PushDevice, error) {
	if s.pushDevices == nil {
		return PushDevice{}, ErrPushDevicesUnavailable
	}
	input.ExpoPushToken = strings.TrimSpace(input.ExpoPushToken)
	input.Platform = strings.ToLower(strings.TrimSpace(input.Platform))
	input.DeviceName = strings.TrimSpace(input.DeviceName)
	input.AppVersion = strings.TrimSpace(input.AppVersion)
	if !validExpoPushToken(input.ExpoPushToken) {
		return PushDevice{}, ErrInvalidPushToken
	}
	if input.Platform != "ios" && input.Platform != "android" {
		return PushDevice{}, ErrInvalidPushPlatform
	}
	return s.pushDevices.RegisterPushDevice(ctx, organizationID, userID, input)
}

func (s *Service) ListPushDevices(ctx context.Context, organizationID, userID string) ([]PushDevice, error) {
	if s.pushDevices == nil {
		return nil, ErrPushDevicesUnavailable
	}
	return s.pushDevices.ListPushDevices(ctx, organizationID, userID)
}

func (s *Service) DeletePushDevice(ctx context.Context, organizationID, userID, deviceID string) error {
	if s.pushDevices == nil {
		return ErrPushDevicesUnavailable
	}
	return s.pushDevices.DeletePushDevice(ctx, organizationID, userID, strings.TrimSpace(deviceID))
}

func (s *Service) QueuePushToUser(ctx context.Context, organizationID, actorUserID string, input PushToUserInput) ([]Notification, error) {
	if s.pushDevices == nil {
		return nil, ErrPushDevicesUnavailable
	}
	input.UserID = strings.TrimSpace(input.UserID)
	input.Topic = strings.TrimSpace(input.Topic)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Body = strings.TrimSpace(input.Body)
	input.ResourceType = strings.TrimSpace(input.ResourceType)
	input.ResourceID = strings.TrimSpace(input.ResourceID)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.UserID == "" {
		return nil, ErrPushUserRequired
	}
	if input.Topic == "" {
		return nil, ErrTopicRequired
	}
	if input.Body == "" {
		return nil, ErrBodyRequired
	}
	if (input.ResourceType == "") != (input.ResourceID == "") {
		return nil, ErrResourcePairRequired
	}
	devices, err := s.pushDevices.ListPushDevices(ctx, organizationID, input.UserID)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, ErrPushRecipientUnavailable
	}
	items := make([]Notification, 0, len(devices))
	for _, device := range devices {
		key := input.IdempotencyKey
		if key != "" {
			key = key + ":" + device.ID
		}
		item, err := s.Enqueue(ctx, organizationID, actorUserID, EnqueueInput{
			Topic: input.Topic, Channel: "push", Recipient: device.ExpoPushToken,
			Subject: input.Subject, Body: input.Body, Payload: input.Payload,
			ResourceType: input.ResourceType, ResourceID: input.ResourceID, IdempotencyKey: key,
		})
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
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
		Topic: "rent.reminder", Channel: input.Channel, Recipient: input.Recipient,
		Subject: subject, Body: body, ResourceType: "rent_obligation", ResourceID: context.ObligationID,
		IdempotencyKey: key,
	})
}

func validChannel(channel string) bool {
	switch channel {
	case "email", "sms", "whatsapp", "webhook", "push":
		return true
	default:
		return false
	}
}

func validExpoPushToken(token string) bool {
	if len(token) < 20 || len(token) > 512 || !strings.HasSuffix(token, "]") {
		return false
	}
	return strings.HasPrefix(token, "ExpoPushToken[") || strings.HasPrefix(token, "ExponentPushToken[")
}
