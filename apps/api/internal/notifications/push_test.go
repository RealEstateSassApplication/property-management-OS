package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakePushRepository struct {
	fakeRepository
	devices    []PushDevice
	registered RegisterPushDeviceInput
	queuedPush []EnqueueInput
	deletedID  string
}

func (f *fakePushRepository) Enqueue(_ context.Context, _, _ string, input EnqueueInput) (Notification, error) {
	f.queuedPush = append(f.queuedPush, input)
	return Notification{ID: "notification-" + input.Recipient, Topic: input.Topic, Channel: input.Channel, Recipient: input.Recipient}, nil
}

func (f *fakePushRepository) RegisterPushDevice(_ context.Context, organizationID, userID string, input RegisterPushDeviceInput) (PushDevice, error) {
	f.registered = input
	return PushDevice{ID: "device-1", OrganizationID: organizationID, UserID: userID, ExpoPushToken: input.ExpoPushToken, Platform: input.Platform}, nil
}

func (f *fakePushRepository) ListPushDevices(_ context.Context, _, _ string) ([]PushDevice, error) {
	return f.devices, nil
}

func (f *fakePushRepository) DeletePushDevice(_ context.Context, _, _, deviceID string) error {
	f.deletedID = deviceID
	return nil
}

func TestRegisterPushDeviceValidatesExpoTokenAndPlatform(t *testing.T) {
	repo := &fakePushRepository{}
	service := NewService(repo)
	if _, err := service.RegisterPushDevice(context.Background(), "org-1", "user-1", RegisterPushDeviceInput{ExpoPushToken: "bad-token", Platform: "ios"}); err != ErrInvalidPushToken {
		t.Fatalf("expected invalid token error, got %v", err)
	}
	if _, err := service.RegisterPushDevice(context.Background(), "org-1", "user-1", RegisterPushDeviceInput{ExpoPushToken: "ExpoPushToken[abcdefghijklmnopqrstuvwxyz]", Platform: "desktop"}); err != ErrInvalidPushPlatform {
		t.Fatalf("expected invalid platform error, got %v", err)
	}
	item, err := service.RegisterPushDevice(context.Background(), "org-1", "user-1", RegisterPushDeviceInput{ExpoPushToken: " ExpoPushToken[abcdefghijklmnopqrstuvwxyz] ", Platform: "IOS", DeviceName: " Phone "})
	if err != nil {
		t.Fatal(err)
	}
	if item.Platform != "ios" || repo.registered.DeviceName != "Phone" {
		t.Fatalf("unexpected registered device: %+v / %+v", item, repo.registered)
	}
}

func TestQueuePushToUserExpandsRegisteredDevices(t *testing.T) {
	repo := &fakePushRepository{devices: []PushDevice{
		{ID: "device-a", ExpoPushToken: "ExpoPushToken[aaaaaaaaaaaaaaaaaaaaaaaaaa]"},
		{ID: "device-b", ExpoPushToken: "ExpoPushToken[bbbbbbbbbbbbbbbbbbbbbbbbbb]"},
	}}
	service := NewService(repo)
	items, err := service.QueuePushToUser(context.Background(), "org-1", "actor-1", PushToUserInput{
		UserID: "user-2", Topic: "maintenance.updated", Subject: "Maintenance updated", Body: "Your request changed.",
		ResourceType: "maintenance_request", ResourceID: "11111111-1111-1111-1111-111111111111", IdempotencyKey: "event-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || len(repo.queuedPush) != 2 {
		t.Fatalf("expected two push notifications, got items=%d queued=%d", len(items), len(repo.queuedPush))
	}
	if repo.queuedPush[0].Channel != "push" || repo.queuedPush[1].Channel != "push" {
		t.Fatalf("expected push channel: %+v", repo.queuedPush)
	}
	if repo.queuedPush[0].IdempotencyKey == repo.queuedPush[1].IdempotencyKey || !strings.HasSuffix(repo.queuedPush[0].IdempotencyKey, ":device-a") || !strings.HasSuffix(repo.queuedPush[1].IdempotencyKey, ":device-b") {
		t.Fatalf("expected per-device idempotency keys: %+v", repo.queuedPush)
	}
}

func TestQueuePushToUserRejectsMissingDevices(t *testing.T) {
	service := NewService(&fakePushRepository{})
	_, err := service.QueuePushToUser(context.Background(), "org-1", "actor-1", PushToUserInput{UserID: "user-2", Topic: "test", Body: "hello"})
	if err != ErrPushRecipientUnavailable {
		t.Fatalf("expected unavailable recipient, got %v", err)
	}
}

func TestExpoPushProviderBuildsResourceDeepLink(t *testing.T) {
	var authorization string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"status":"ok","id":"ticket-1"}}`))
	}))
	defer server.Close()

	provider := NewExpoPushProvider(server.URL, "push-secret")
	err := provider.Send(context.Background(), Notification{
		Topic: "inspection.updated", Recipient: "ExpoPushToken[abcdefghijklmnopqrstuvwxyz]", Subject: "Inspection updated", Body: "Open the inspection.",
		ResourceType: "inspection", ResourceID: "22222222-2222-2222-2222-222222222222", Payload: json.RawMessage(`{"custom":"value"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if authorization != "Bearer push-secret" {
		t.Fatalf("unexpected authorization header: %q", authorization)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing push data payload: %+v", payload)
	}
	if data["custom"] != "value" || data["resourceType"] != "inspection" || !strings.HasPrefix(data["url"].(string), "propertyos://open?") {
		t.Fatalf("unexpected push data: %+v", data)
	}
}

type recordingProvider struct{ called int }

func (p *recordingProvider) Send(context.Context, Notification) error {
	p.called++
	return nil
}

func TestRoutingProviderUsesChannelOverride(t *testing.T) {
	fallback := &recordingProvider{}
	push := &recordingProvider{}
	provider := NewRoutingProvider(fallback, map[string]DeliveryProvider{"push": push})
	if err := provider.Send(context.Background(), Notification{Channel: "push"}); err != nil {
		t.Fatal(err)
	}
	if err := provider.Send(context.Background(), Notification{Channel: "email"}); err != nil {
		t.Fatal(err)
	}
	if push.called != 1 || fallback.called != 1 {
		t.Fatalf("unexpected routing calls push=%d fallback=%d", push.called, fallback.called)
	}
}
