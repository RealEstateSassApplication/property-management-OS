package propertyosclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/leases"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/notifications"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/properties"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/rent"
)

type Client struct {
	baseURL      string
	organization string
	accessToken  string
	userID       string
	httpClient   *http.Client
}

func NewFromEnv() (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PROPERTY_OS_API_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	organization := strings.TrimSpace(os.Getenv("PROPERTY_OS_ORGANIZATION_ID"))
	accessToken := strings.TrimSpace(os.Getenv("PROPERTY_OS_ACCESS_TOKEN"))
	userID := strings.TrimSpace(os.Getenv("PROPERTY_OS_USER_ID"))
	if organization == "" {
		return nil, errors.New("PROPERTY_OS_ORGANIZATION_ID is required")
	}
	if accessToken == "" && userID == "" {
		return nil, errors.New("PROPERTY_OS_ACCESS_TOKEN is required outside development; PROPERTY_OS_USER_ID may be used for a local API")
	}
	return &Client{baseURL: baseURL, organization: organization, accessToken: accessToken, userID: userID, httpClient: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *Client) ListProperties(ctx context.Context) ([]properties.Property, error) {
	return getList[properties.Property](ctx, c, "/api/v1/properties")
}
func (c *Client) ListRentObligations(ctx context.Context) ([]rent.Obligation, error) {
	return getList[rent.Obligation](ctx, c, "/api/v1/rent/obligations")
}
func (c *Client) ListLeases(ctx context.Context) ([]leases.Lease, error) {
	return getList[leases.Lease](ctx, c, "/api/v1/leases")
}
func (c *Client) ListMaintenanceRequests(ctx context.Context) ([]maintenance.Request, error) {
	return getList[maintenance.Request](ctx, c, "/api/v1/maintenance/requests")
}
func (c *Client) ListNotifications(ctx context.Context) ([]notifications.Notification, error) {
	return getList[notifications.Notification](ctx, c, "/api/v1/notifications")
}
func (c *Client) QueueRentReminder(ctx context.Context, input notifications.RentReminderInput) (notifications.Notification, error) {
	return postOne[notifications.RentReminderInput, notifications.Notification](ctx, c, "/api/v1/notifications/rent-reminders", input)
}
func (c *Client) CreateMaintenanceRequest(ctx context.Context, input maintenance.CreateRequestInput) (maintenance.Request, error) {
	return postOne[maintenance.CreateRequestInput, maintenance.Request](ctx, c, "/api/v1/maintenance/requests", input)
}

func getList[T any](ctx context.Context, client *Client, path string) ([]T, error) {
	var response struct{ Data []T `json:"data"` }
	if err := client.do(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func postOne[I any, O any](ctx context.Context, client *Client, path string, input I) (O, error) {
	var response struct{ Data O `json:"data"` }
	var zero O
	if err := client.do(ctx, http.MethodPost, path, input, &response); err != nil {
		return zero, err
	}
	return response.Data, nil
}

func (c *Client) do(ctx context.Context, method, path string, input, output any) error {
	var body *bytes.Reader
	if input == nil {
		body = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Organization-ID", c.organization)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	} else {
		req.Header.Set("X-User-ID", c.userID)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var problem struct{ Error struct{ Message string `json:"message"` } `json:"error"` }
		_ = json.NewDecoder(res.Body).Decode(&problem)
		if problem.Error.Message == "" {
			problem.Error.Message = res.Status
		}
		return fmt.Errorf("property OS API: %s", problem.Error.Message)
	}
	return json.NewDecoder(res.Body).Decode(output)
}
