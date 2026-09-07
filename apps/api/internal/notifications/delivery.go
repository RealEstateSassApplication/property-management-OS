package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type DeliveryProvider interface {
	Send(ctx context.Context, notification Notification) error
}

type LogProvider struct{ logger *slog.Logger }

func NewLogProvider(logger *slog.Logger) *LogProvider { return &LogProvider{logger: logger} }
func (p *LogProvider) Send(_ context.Context, n Notification) error {
	p.logger.Info("notification delivery", "id", n.ID, "topic", n.Topic, "channel", n.Channel, "recipient", n.Recipient, "subject", n.Subject)
	return nil
}

type WebhookProvider struct {
	url    string
	token  string
	client *http.Client
}

func NewWebhookProvider(url, token string) *WebhookProvider {
	return &WebhookProvider{url: strings.TrimSpace(url), token: strings.TrimSpace(token), client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *WebhookProvider) Send(ctx context.Context, n Notification) error {
	payload, err := json.Marshal(map[string]any{"event": "property_os.notification", "notification": n})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("notification webhook returned %d", res.StatusCode)
	}
	return nil
}
