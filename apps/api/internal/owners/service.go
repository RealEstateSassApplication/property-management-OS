package owners

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type Repository interface {
	List(ctx context.Context, organizationID string) ([]Owner, error)
	Get(ctx context.Context, organizationID, ownerID string) (Owner, error)
	Create(ctx context.Context, organizationID string, input CreateOwnerInput) (Owner, error)
	ListInterests(ctx context.Context, organizationID string) ([]OwnershipInterest, error)
	CreateInterest(ctx context.Context, organizationID string, input CreateInterestInput) (OwnershipInterest, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, organizationID string) ([]Owner, error) {
	return s.repository.List(ctx, organizationID)
}

func (s *Service) Get(ctx context.Context, organizationID, ownerID string) (Owner, error) {
	return s.repository.Get(ctx, organizationID, strings.TrimSpace(ownerID))
}

func (s *Service) Create(ctx context.Context, organizationID string, input CreateOwnerInput) (Owner, error) {
	input.LegalName = strings.TrimSpace(input.LegalName)
	input.OwnerType = strings.ToLower(strings.TrimSpace(input.OwnerType))
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.OwnerType == "" {
		input.OwnerType = "individual"
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.LegalName == "" {
		return Owner{}, fmt.Errorf("legalName is required")
	}
	if input.OwnerType != "individual" && input.OwnerType != "company" {
		return Owner{}, fmt.Errorf("ownerType must be individual or company")
	}
	if input.Status != "active" && input.Status != "inactive" {
		return Owner{}, fmt.Errorf("status must be active or inactive")
	}
	if input.Email != "" {
		if _, err := mail.ParseAddress(input.Email); err != nil {
			return Owner{}, fmt.Errorf("email is invalid")
		}
	}
	return s.repository.Create(ctx, organizationID, input)
}

func (s *Service) ListInterests(ctx context.Context, organizationID string) ([]OwnershipInterest, error) {
	return s.repository.ListInterests(ctx, organizationID)
}

func (s *Service) CreateInterest(ctx context.Context, organizationID string, input CreateInterestInput) (OwnershipInterest, error) {
	input.OwnerID = strings.TrimSpace(input.OwnerID)
	input.PropertyID = strings.TrimSpace(input.PropertyID)
	input.EffectiveFrom = strings.TrimSpace(input.EffectiveFrom)
	if input.OwnerID == "" || input.PropertyID == "" {
		return OwnershipInterest{}, fmt.Errorf("ownerId and propertyId are required")
	}
	if input.OwnershipBps < 1 || input.OwnershipBps > 10000 {
		return OwnershipInterest{}, fmt.Errorf("ownershipBps must be between 1 and 10000")
	}
	if input.EffectiveFrom == "" {
		input.EffectiveFrom = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", input.EffectiveFrom); err != nil {
		return OwnershipInterest{}, fmt.Errorf("effectiveFrom must be YYYY-MM-DD")
	}
	return s.repository.CreateInterest(ctx, organizationID, input)
}
