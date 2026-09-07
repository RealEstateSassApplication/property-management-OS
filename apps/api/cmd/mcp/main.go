package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/notifications"
	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/propertyosclient"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EmptyInput struct{}

type PortfolioSnapshotOutput struct {
	PropertyCount        int              `json:"propertyCount"`
	OverdueObligations   int              `json:"overdueObligations"`
	OutstandingByCurrency map[string]int64 `json:"outstandingByCurrencyMinor"`
	OpenMaintenance     int              `json:"openMaintenance"`
	LeasesExpiring60Days int              `json:"leasesExpiring60Days"`
	QueuedNotifications int              `json:"queuedNotifications"`
}

type ArrearsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"maximum number of overdue obligations to return; defaults to 50"`
}
type ArrearsItem struct {
	ObligationID string `json:"obligationId"`
	Tenant       string `json:"tenant"`
	Property     string `json:"property"`
	Unit         string `json:"unit"`
	DueDate      string `json:"dueDate"`
	BalanceMinor int64  `json:"balanceMinor"`
	Currency     string `json:"currency"`
}
type ArrearsOutput struct{ Items []ArrearsItem `json:"items"` }

type LeaseExpiryInput struct {
	Days int `json:"days,omitempty" jsonschema:"look-ahead window in days; defaults to 60 and is capped at 365"`
}
type LeaseExpiryItem struct {
	LeaseID    string `json:"leaseId"`
	Reference  string `json:"reference"`
	Tenant     string `json:"tenant"`
	Property   string `json:"property"`
	Unit       string `json:"unit"`
	EndDate    string `json:"endDate"`
}
type LeaseExpiryOutput struct{ Items []LeaseExpiryItem `json:"items"` }

type MaintenanceQueueInput struct {
	Priority string `json:"priority,omitempty" jsonschema:"optional priority filter: low, normal, high, or emergency"`
}
type MaintenanceQueueItem struct {
	RequestID string `json:"requestId"`
	Title     string `json:"title"`
	Property  string `json:"property"`
	Unit      string `json:"unit,omitempty"`
	Priority  string `json:"priority"`
	Status    string `json:"status"`
}
type MaintenanceQueueOutput struct{ Items []MaintenanceQueueItem `json:"items"` }

type RentReminderInput struct {
	ObligationID string `json:"obligationId" jsonschema:"rent obligation UUID"`
	Channel      string `json:"channel" jsonschema:"email, sms, or whatsapp"`
	Recipient    string `json:"recipient" jsonschema:"destination email address or phone number"`
}
type CreateMaintenanceInput struct {
	PropertyID  string `json:"propertyId" jsonschema:"property UUID"`
	UnitID      string `json:"unitId,omitempty" jsonschema:"optional unit UUID"`
	TenantID    string `json:"tenantId,omitempty" jsonschema:"optional tenant UUID; API verifies active occupancy"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category" jsonschema:"plumbing, electrical, hvac, appliance, structural, cleaning, security, or other"`
	Priority    string `json:"priority" jsonschema:"low, normal, high, or emergency"`
}
type MutationOutput struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	client, err := propertyosclient.NewFromEnv()
	if err != nil {
		log.New(os.Stderr, "property-os-mcp: ", 0).Fatal(err)
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "property-management-os", Version: "v0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "portfolio_snapshot", Description: "Read a concise operational snapshot of the authenticated Property OS organization."}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, PortfolioSnapshotOutput, error) {
		properties, err := client.ListProperties(ctx)
		if err != nil { return nil, PortfolioSnapshotOutput{}, err }
		obligations, err := client.ListRentObligations(ctx)
		if err != nil { return nil, PortfolioSnapshotOutput{}, err }
		leases, err := client.ListLeases(ctx)
		if err != nil { return nil, PortfolioSnapshotOutput{}, err }
		requests, err := client.ListMaintenanceRequests(ctx)
		if err != nil { return nil, PortfolioSnapshotOutput{}, err }
		notificationsList, err := client.ListNotifications(ctx)
		if err != nil { return nil, PortfolioSnapshotOutput{}, err }
		out := PortfolioSnapshotOutput{PropertyCount: len(properties), OutstandingByCurrency: map[string]int64{}}
		for _, obligation := range obligations {
			if obligation.State == "overdue" && obligation.BalanceMinor > 0 { out.OverdueObligations++ }
			if obligation.BalanceMinor > 0 { out.OutstandingByCurrency[obligation.Currency] += obligation.BalanceMinor }
		}
		cutoff := time.Now().UTC().AddDate(0, 0, 60)
		for _, lease := range leases {
			end, parseErr := time.Parse("2006-01-02", lease.EndDate)
			if parseErr == nil && lease.Status == "active" && !end.Before(time.Now().UTC()) && !end.After(cutoff) { out.LeasesExpiring60Days++ }
		}
		for _, request := range requests { if request.Status != "resolved" && request.Status != "cancelled" { out.OpenMaintenance++ } }
		for _, item := range notificationsList { if item.Status == "pending" || item.Status == "retry" || item.Status == "processing" { out.QueuedNotifications++ } }
		return nil, out, nil
	})

	mcp.AddTool(server, &mcp.Tool{Name: "list_overdue_rent", Description: "List overdue rent obligations with outstanding balances. Read-only."}, func(ctx context.Context, _ *mcp.CallToolRequest, input ArrearsInput) (*mcp.CallToolResult, ArrearsOutput, error) {
		items, err := client.ListRentObligations(ctx)
		if err != nil { return nil, ArrearsOutput{}, err }
		limit := input.Limit
		if limit <= 0 || limit > 200 { limit = 50 }
		out := ArrearsOutput{Items: make([]ArrearsItem, 0)}
		for _, item := range items {
			if item.State != "overdue" || item.BalanceMinor <= 0 { continue }
			out.Items = append(out.Items, ArrearsItem{ObligationID: item.ID, Tenant: item.PrimaryTenantName, Property: item.PropertyName, Unit: item.UnitLabel, DueDate: item.DueDate, BalanceMinor: item.BalanceMinor, Currency: item.Currency})
			if len(out.Items) >= limit { break }
		}
		return nil, out, nil
	})

	mcp.AddTool(server, &mcp.Tool{Name: "list_expiring_leases", Description: "List active leases ending within a requested window. Read-only."}, func(ctx context.Context, _ *mcp.CallToolRequest, input LeaseExpiryInput) (*mcp.CallToolResult, LeaseExpiryOutput, error) {
		items, err := client.ListLeases(ctx)
		if err != nil { return nil, LeaseExpiryOutput{}, err }
		days := input.Days
		if days <= 0 { days = 60 }
		if days > 365 { days = 365 }
		now := time.Now().UTC()
		cutoff := now.AddDate(0, 0, days)
		out := LeaseExpiryOutput{Items: make([]LeaseExpiryItem, 0)}
		for _, item := range items {
			end, parseErr := time.Parse("2006-01-02", item.EndDate)
			if parseErr != nil || item.Status != "active" || end.Before(now) || end.After(cutoff) { continue }
			out.Items = append(out.Items, LeaseExpiryItem{LeaseID: item.ID, Reference: item.ReferenceCode, Tenant: item.PrimaryTenantName, Property: item.PropertyName, Unit: item.UnitLabel, EndDate: item.EndDate})
		}
		sort.Slice(out.Items, func(i, j int) bool { return out.Items[i].EndDate < out.Items[j].EndDate })
		return nil, out, nil
	})

	mcp.AddTool(server, &mcp.Tool{Name: "list_open_maintenance", Description: "List unresolved maintenance requests, optionally filtered by priority. Read-only."}, func(ctx context.Context, _ *mcp.CallToolRequest, input MaintenanceQueueInput) (*mcp.CallToolResult, MaintenanceQueueOutput, error) {
		items, err := client.ListMaintenanceRequests(ctx)
		if err != nil { return nil, MaintenanceQueueOutput{}, err }
		priority := strings.ToLower(strings.TrimSpace(input.Priority))
		out := MaintenanceQueueOutput{Items: make([]MaintenanceQueueItem, 0)}
		for _, item := range items {
			if item.Status == "resolved" || item.Status == "cancelled" || (priority != "" && item.Priority != priority) { continue }
			out.Items = append(out.Items, MaintenanceQueueItem{RequestID: item.ID, Title: item.Title, Property: item.PropertyName, Unit: item.UnitLabel, Priority: item.Priority, Status: item.Status})
		}
		return nil, out, nil
	})

	mcp.AddTool(server, &mcp.Tool{Name: "queue_rent_reminder", Description: "Mutating action. Queue a server-authored reminder for an outstanding rent obligation. The API validates the obligation, derives the balance/due date, enforces RBAC, and deduplicates same-day reminders."}, func(ctx context.Context, _ *mcp.CallToolRequest, input RentReminderInput) (*mcp.CallToolResult, MutationOutput, error) {
		item, err := client.QueueRentReminder(ctx, notifications.RentReminderInput{ObligationID: input.ObligationID, Channel: input.Channel, Recipient: input.Recipient})
		if err != nil { return nil, MutationOutput{}, err }
		return nil, MutationOutput{ID: item.ID, Status: item.Status, Message: "rent reminder queued for delivery"}, nil
	})

	mcp.AddTool(server, &mcp.Tool{Name: "create_maintenance_request", Description: "Mutating action. Create a maintenance request through the normal Property OS API. Tenant/unit occupancy and RBAC are enforced by the backend."}, func(ctx context.Context, _ *mcp.CallToolRequest, input CreateMaintenanceInput) (*mcp.CallToolResult, MutationOutput, error) {
		item, err := client.CreateMaintenanceRequest(ctx, maintenance.CreateRequestInput{PropertyID: input.PropertyID, UnitID: input.UnitID, TenantID: input.TenantID, Title: input.Title, Description: input.Description, Category: input.Category, Priority: input.Priority})
		if err != nil { return nil, MutationOutput{}, err }
		return nil, MutationOutput{ID: item.ID, Status: item.Status, Message: fmt.Sprintf("maintenance request created: %s", item.Title)}, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.New(os.Stderr, "property-os-mcp: ", 0).Fatal(err)
	}
}
