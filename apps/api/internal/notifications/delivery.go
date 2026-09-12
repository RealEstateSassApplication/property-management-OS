package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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

type RoutingProvider struct {
	fallback  DeliveryProvider
	overrides map[string]DeliveryProvider
}

func NewRoutingProvider(fallback DeliveryProvider, overrides map[string]DeliveryProvider) *RoutingProvider {
	return &RoutingProvider{fallback: fallback, overrides: overrides}
}

func (p *RoutingProvider) Send(ctx context.Context, n Notification) error {
	if provider := p.overrides[strings.ToLower(strings.TrimSpace(n.Channel))]; provider != nil {
		return provider.Send(ctx, n)
	}
	if p.fallback == nil {
		return fmt.Errorf("no delivery provider configured for channel %q", n.Channel)
	}
	return p.fallback.Send(ctx, n)
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

type ExpoPushProvider struct {
	endpoint    string
	accessToken string
	client      *http.Client
}

func NewExpoPushProvider(endpoint, accessToken string) *ExpoPushProvider {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = "https://exp.host/--/api/v2/push/send"
	}
	return &ExpoPushProvider{endpoint: endpoint, accessToken: strings.TrimSpace(accessToken), client: &http.Client{Timeout: 15 * time.Second}}
}

func (p *ExpoPushProvider) Send(ctx context.Context, n Notification) error {
	data := map[string]any{}
	if len(n.Payload) > 0 {
		_ = json.Unmarshal(n.Payload, &data)
	}
	if n.ResourceType != "" && n.ResourceID != "" {
		data["resourceType"] = n.ResourceType
		data["resourceId"] = n.ResourceID
		data["url"] = "propertyos://open?resourceType=" + url.QueryEscape(n.ResourceType) + "&resourceId=" + url.QueryEscape(n.ResourceID)
	}
	title := strings.TrimSpace(n.Subject)
	if title == "" {
		title = strings.TrimSpace(n.Topic)
	}
	payload, err := json.Marshal(map[string]any{
		"to": n.Recipient, "title": title, "body": n.Body, "data": data, "sound": "default",
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if p.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.accessToken)
	}
	res, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Expo push service returned %d", res.StatusCode)
	}
	var result struct {
		Data struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("could not decode Expo push response: %w", err)
	}
	if result.Data.Status != "ok" {
		if result.Data.Message == "" {
			result.Data.Message = "push ticket was not accepted"
		}
		return fmt.Errorf("Expo push delivery failed: %s", result.Data.Message)
	}
	return nil
}
