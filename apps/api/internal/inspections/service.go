package inspections

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	List(ctx context.Context, organizationID string) ([]Inspection, error)
	Get(ctx context.Context, organizationID, inspectionID string) (Inspection, error)
	ListItems(ctx context.Context, organizationID, inspectionID string) ([]Item, error)
	Create(ctx context.Context, organizationID, actorUserID string, input CreateInspectionInput) (Inspection, error)
	CreateItem(ctx context.Context, organizationID string, input CreateItemInput) (Item, error)
	Complete(ctx context.Context, organizationID, actorUserID string, input CompleteInspectionInput) (Inspection, error)
	Acknowledge(ctx context.Context, organizationID, actorUserID string, input AcknowledgeInspectionInput) (Inspection, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, organizationID string) ([]Inspection, error) {
	return s.repository.List(ctx, organizationID)
}

func (s *Service) Get(ctx context.Context, organizationID, inspectionID string) (Inspection, error) {
	inspectionID = strings.TrimSpace(inspectionID)
	if inspectionID == "" {
		return Inspection{}, fmt.Errorf("inspectionId is required")
	}
	return s.repository.Get(ctx, organizationID, inspectionID)
}

func (s *Service) ListItems(ctx context.Context, organizationID, inspectionID string) ([]Item, error) {
	inspectionID = strings.TrimSpace(inspectionID)
	if inspectionID == "" {
		return nil, fmt.Errorf("inspectionId is required")
	}
	return s.repository.ListItems(ctx, organizationID, inspectionID)
}

func (s *Service) Create(ctx context.Context, organizationID, actorUserID string, input CreateInspectionInput) (Inspection, error) {
	input.TenancyID = strings.TrimSpace(input.TenancyID)
	input.InspectionType = strings.ToLower(strings.TrimSpace(input.InspectionType))
	input.ScheduledFor = strings.TrimSpace(input.ScheduledFor)
	input.Summary = strings.TrimSpace(input.Summary)
	if input.TenancyID == "" {
		return Inspection{}, fmt.Errorf("tenancyId is required")
	}
	if !contains([]string{"move_in", "move_out", "periodic"}, input.InspectionType) {
		return Inspection{}, ErrInvalidInspectionType
	}
	if input.ScheduledFor != "" {
		if _, err := time.Parse(time.RFC3339, input.ScheduledFor); err != nil {
			return Inspection{}, fmt.Errorf("scheduledFor must use RFC3339")
		}
	}
	return s.repository.Create(ctx, organizationID, actorUserID, input)
}

func (s *Service) CreateItem(ctx context.Context, organizationID string, input CreateItemInput) (Item, error) {
	input.InspectionID = strings.TrimSpace(input.InspectionID)
	input.Area = strings.TrimSpace(input.Area)
	input.ItemName = strings.TrimSpace(input.ItemName)
	input.Condition = strings.ToLower(strings.TrimSpace(input.Condition))
	input.Notes = strings.TrimSpace(input.Notes)
	input.EvidenceDocumentID = strings.TrimSpace(input.EvidenceDocumentID)
	if input.InspectionID == "" || input.Area == "" || input.ItemName == "" {
		return Item{}, ErrInvalidInspectionItem
	}
	if !contains([]string{"good", "fair", "poor", "damaged", "not_applicable"}, input.Condition) {
		return Item{}, ErrInvalidInspectionItem
	}
	return s.repository.CreateItem(ctx, organizationID, input)
}

func (s *Service) Complete(ctx context.Context, organizationID, actorUserID string, input CompleteInspectionInput) (Inspection, error) {
	input.InspectionID = strings.TrimSpace(input.InspectionID)
	input.Summary = strings.TrimSpace(input.Summary)
	if input.InspectionID == "" {
		return Inspection{}, fmt.Errorf("inspectionId is required")
	}
	return s.repository.Complete(ctx, organizationID, actorUserID, input)
}

func (s *Service) Acknowledge(ctx context.Context, organizationID, actorUserID string, input AcknowledgeInspectionInput) (Inspection, error) {
	input.InspectionID = strings.TrimSpace(input.InspectionID)
	if input.InspectionID == "" {
		return Inspection{}, fmt.Errorf("inspectionId is required")
	}
	return s.repository.Acknowledge(ctx, organizationID, actorUserID, input)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
