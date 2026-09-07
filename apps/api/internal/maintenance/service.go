package maintenance

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	ListVendors(ctx context.Context, organizationID string) ([]Vendor, error)
	CreateVendor(ctx context.Context, organizationID string, input CreateVendorInput) (Vendor, error)
	ListRequests(ctx context.Context, organizationID string) ([]Request, error)
	CreateRequest(ctx context.Context, organizationID, actorUserID string, input CreateRequestInput) (Request, error)
	UpdateRequestStatus(ctx context.Context, organizationID, requestID, status string) (Request, error)
	ListWorkOrders(ctx context.Context, organizationID string) ([]WorkOrder, error)
	CreateWorkOrder(ctx context.Context, organizationID string, input CreateWorkOrderInput, scheduledFor *time.Time) (WorkOrder, error)
	UpdateWorkOrderStatus(ctx context.Context, organizationID, workOrderID, status string) (WorkOrder, error)
	ListQuotes(ctx context.Context, organizationID string) ([]Quote, error)
	CreateQuote(ctx context.Context, organizationID string, input CreateQuoteInput) (Quote, error)
	DecideQuote(ctx context.Context, organizationID, quoteID, actorUserID, decision string) (Quote, error)
	ListEvidence(ctx context.Context, organizationID string) ([]Evidence, error)
	CreateEvidence(ctx context.Context, organizationID, actorUserID string, input CreateEvidenceInput) (Evidence, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListVendors(ctx context.Context, organizationID string) ([]Vendor, error) {
	return s.repository.ListVendors(ctx, organizationID)
}

func (s *Service) CreateVendor(ctx context.Context, organizationID string, input CreateVendorInput) (Vendor, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Trade = strings.ToLower(strings.TrimSpace(input.Trade))
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "active"
	}
	if input.Name == "" {
		return Vendor{}, fmt.Errorf("vendor name is required")
	}
	if !validTrade(input.Trade) {
		return Vendor{}, fmt.Errorf("unsupported vendor trade")
	}
	if input.Status != "active" && input.Status != "inactive" {
		return Vendor{}, fmt.Errorf("vendor status must be active or inactive")
	}
	return s.repository.CreateVendor(ctx, organizationID, input)
}

func (s *Service) ListRequests(ctx context.Context, organizationID string) ([]Request, error) {
	return s.repository.ListRequests(ctx, organizationID)
}

func (s *Service) CreateRequest(ctx context.Context, organizationID, actorUserID string, input CreateRequestInput) (Request, error) {
	input.PropertyID = strings.TrimSpace(input.PropertyID)
	input.UnitID = strings.TrimSpace(input.UnitID)
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	input.Priority = strings.ToLower(strings.TrimSpace(input.Priority))
	if input.Priority == "" {
		input.Priority = "normal"
	}
	if input.PropertyID == "" || input.Title == "" || input.Description == "" {
		return Request{}, fmt.Errorf("propertyId, title and description are required")
	}
	if !validCategory(input.Category) {
		return Request{}, fmt.Errorf("unsupported maintenance category")
	}
	if !validPriority(input.Priority) {
		return Request{}, fmt.Errorf("unsupported maintenance priority")
	}
	return s.repository.CreateRequest(ctx, organizationID, strings.TrimSpace(actorUserID), input)
}

func (s *Service) UpdateRequestStatus(ctx context.Context, organizationID, requestID string, input UpdateRequestStatusInput) (Request, error) {
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		return Request{}, fmt.Errorf("status is required")
	}
	return s.repository.UpdateRequestStatus(ctx, organizationID, strings.TrimSpace(requestID), status)
}

func (s *Service) ListWorkOrders(ctx context.Context, organizationID string) ([]WorkOrder, error) {
	return s.repository.ListWorkOrders(ctx, organizationID)
}

func (s *Service) CreateWorkOrder(ctx context.Context, organizationID string, input CreateWorkOrderInput) (WorkOrder, error) {
	input.MaintenanceRequestID = strings.TrimSpace(input.MaintenanceRequestID)
	input.VendorID = strings.TrimSpace(input.VendorID)
	input.AssignedUserID = strings.TrimSpace(input.AssignedUserID)
	input.Summary = strings.TrimSpace(input.Summary)
	input.ScheduledFor = strings.TrimSpace(input.ScheduledFor)
	if input.MaintenanceRequestID == "" || input.Summary == "" {
		return WorkOrder{}, fmt.Errorf("maintenanceRequestId and summary are required")
	}
	var scheduledFor *time.Time
	if input.ScheduledFor != "" {
		parsed, err := time.Parse(time.RFC3339, input.ScheduledFor)
		if err != nil {
			return WorkOrder{}, fmt.Errorf("scheduledFor must be RFC3339")
		}
		scheduledFor = &parsed
	}
	return s.repository.CreateWorkOrder(ctx, organizationID, input, scheduledFor)
}

func (s *Service) UpdateWorkOrderStatus(ctx context.Context, organizationID, workOrderID string, input UpdateWorkOrderStatusInput) (WorkOrder, error) {
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		return WorkOrder{}, fmt.Errorf("status is required")
	}
	return s.repository.UpdateWorkOrderStatus(ctx, organizationID, strings.TrimSpace(workOrderID), status)
}

func (s *Service) ListQuotes(ctx context.Context, organizationID string) ([]Quote, error) {
	return s.repository.ListQuotes(ctx, organizationID)
}

func (s *Service) CreateQuote(ctx context.Context, organizationID string, input CreateQuoteInput) (Quote, error) {
	input.WorkOrderID = strings.TrimSpace(input.WorkOrderID)
	input.VendorID = strings.TrimSpace(input.VendorID)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.ScopeSummary = strings.TrimSpace(input.ScopeSummary)
	if input.WorkOrderID == "" || input.VendorID == "" || input.ScopeSummary == "" {
		return Quote{}, fmt.Errorf("workOrderId, vendorId and scopeSummary are required")
	}
	if input.AmountMinor <= 0 {
		return Quote{}, fmt.Errorf("quote amount must be positive")
	}
	if len(input.Currency) != 3 {
		return Quote{}, fmt.Errorf("currency must be a 3-letter code")
	}
	return s.repository.CreateQuote(ctx, organizationID, input)
}

func (s *Service) DecideQuote(ctx context.Context, organizationID, quoteID, actorUserID string, input DecideQuoteInput) (Quote, error) {
	decision := strings.ToLower(strings.TrimSpace(input.Decision))
	if decision != "approve" && decision != "reject" {
		return Quote{}, fmt.Errorf("decision must be approve or reject")
	}
	return s.repository.DecideQuote(ctx, organizationID, strings.TrimSpace(quoteID), strings.TrimSpace(actorUserID), decision)
}

func (s *Service) ListEvidence(ctx context.Context, organizationID string) ([]Evidence, error) {
	return s.repository.ListEvidence(ctx, organizationID)
}

func (s *Service) CreateEvidence(ctx context.Context, organizationID, actorUserID string, input CreateEvidenceInput) (Evidence, error) {
	input.WorkOrderID = strings.TrimSpace(input.WorkOrderID)
	input.EvidenceType = strings.ToLower(strings.TrimSpace(input.EvidenceType))
	input.Note = strings.TrimSpace(input.Note)
	input.StorageKey = strings.TrimSpace(input.StorageKey)
	if input.WorkOrderID == "" {
		return Evidence{}, fmt.Errorf("workOrderId is required")
	}
	if !validEvidenceType(input.EvidenceType) {
		return Evidence{}, fmt.Errorf("unsupported evidence type")
	}
	if input.Note == "" && input.StorageKey == "" {
		return Evidence{}, fmt.Errorf("evidence requires a note or storageKey")
	}
	return s.repository.CreateEvidence(ctx, organizationID, strings.TrimSpace(actorUserID), input)
}

func validTrade(value string) bool {
	switch value {
	case "plumbing", "electrical", "hvac", "appliance", "structural", "cleaning", "security", "general", "other":
		return true
	default:
		return false
	}
}

func validCategory(value string) bool {
	switch value {
	case "plumbing", "electrical", "hvac", "appliance", "structural", "cleaning", "security", "other":
		return true
	default:
		return false
	}
}

func validPriority(value string) bool {
	return value == "low" || value == "normal" || value == "high" || value == "emergency"
}

func validEvidenceType(value string) bool {
	switch value {
	case "note", "photo", "invoice", "receipt", "other":
		return true
	default:
		return false
	}
}
