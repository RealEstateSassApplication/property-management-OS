package notifications

import (
	"context"
	"testing"
	"time"
)

type fakeRepository struct {
	reminder RentReminderContext
	queued   EnqueueInput
}

func (f *fakeRepository) List(context.Context, string) ([]Notification, error) { return nil, nil }
func (f *fakeRepository) Enqueue(_ context.Context, _, _ string, input EnqueueInput) (Notification, error) {
	f.queued = input
	return Notification{ID: "notification-1", Topic: input.Topic, Body: input.Body}, nil
}
func (f *fakeRepository) GetRentReminderContext(context.Context, string, string) (RentReminderContext, error) {
	return f.reminder, nil
}
func (f *fakeRepository) ClaimBatch(context.Context, string, int) ([]Notification, error) {
	return nil, nil
}
func (f *fakeRepository) MarkDelivered(context.Context, string) error { return nil }
func (f *fakeRepository) MarkFailed(context.Context, string, string, time.Time, bool) error {
	return nil
}

func TestQueueRentReminderBuildsServerAuthoritativeMessage(t *testing.T) {
	repo := &fakeRepository{reminder: RentReminderContext{ObligationID: "obligation-1", TenantName: "Maya", PropertyName: "Lake House", UnitLabel: "A1", Period: "2026-09", DueDate: "2026-09-05", BalanceMinor: 5000000, Currency: "LKR", State: "overdue"}}
	service := NewService(repo)
	_, err := service.QueueRentReminder(context.Background(), "org-1", "user-1", RentReminderInput{ObligationID: "obligation-1", Channel: "email", Recipient: "maya@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.queued.Topic != "rent.reminder" || repo.queued.ResourceID != "obligation-1" || repo.queued.Body == "" || repo.queued.IdempotencyKey == "" {
		t.Fatalf("unexpected queued reminder: %+v", repo.queued)
	}
}

func TestEnqueueRejectsPartialResourcePair(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Enqueue(context.Background(), "org-1", "user-1", EnqueueInput{Topic: "test", Channel: "email", Recipient: "a@example.com", Body: "hello", ResourceType: "lease"})
	if err != ErrResourcePairRequired {
		t.Fatalf("expected ErrResourcePairRequired, got %v", err)
	}
}
