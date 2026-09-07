package tenants

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

var ErrNotFound = errors.New("tenant not found")

type Repository interface {
	List(ctx context.Context, organizationID string) ([]Tenant, error)
	Get(ctx context.Context, organizationID, tenantID string) (Tenant, error)
	Create(ctx context.Context, organizationID string, input CreateInput) (Tenant, error)
	Update(ctx context.Context, organizationID, tenantID string, input UpdateInput) (Tenant, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, organizationID string) ([]Tenant, error) {
	return s.repository.List(ctx, organizationID)
}

func (s *Service) Get(ctx context.Context, organizationID, tenantID string) (Tenant, error) {
	return s.repository.Get(ctx, organizationID, tenantID)
}

func (s *Service) Create(ctx context.Context, organizationID string, input CreateInput) (Tenant, error) {
	normalized, err := normalizeCreate(input)
	if err != nil {
		return Tenant{}, err
	}
	return s.repository.Create(ctx, organizationID, normalized)
}

func (s *Service) Update(ctx context.Context, organizationID, tenantID string, input UpdateInput) (Tenant, error) {
	if input.LegalName != nil {
		value := strings.TrimSpace(*input.LegalName)
		if value == "" {
			return Tenant{}, fmt.Errorf("legalName is required")
		}
		input.LegalName = &value
	}
	if input.Email != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Email))
		if value != "" {
			if _, err := mail.ParseAddress(value); err != nil {
				return Tenant{}, fmt.Errorf("email is invalid")
			}
		}
		input.Email = &value
	}
	if input.Phone != nil {
		value := strings.TrimSpace(*input.Phone)
		input.Phone = &value
	}
	if input.Status != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Status))
		if !validStatus(value) {
			return Tenant{}, fmt.Errorf("status must be one of prospect, active, former, blocked")
		}
		input.Status = &value
	}
	return s.repository.Update(ctx, organizationID, tenantID, input)
}

func normalizeCreate(input CreateInput) (CreateInput, error) {
	input.LegalName = strings.TrimSpace(input.LegalName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "prospect"
	}
	if input.LegalName == "" {
		return CreateInput{}, fmt.Errorf("legalName is required")
	}
	if input.Email != "" {
		if _, err := mail.ParseAddress(input.Email); err != nil {
			return CreateInput{}, fmt.Errorf("email is invalid")
		}
	}
	if !validStatus(input.Status) {
		return CreateInput{}, fmt.Errorf("status must be one of prospect, active, former, blocked")
	}
	return input, nil
}

func validStatus(status string) bool {
	switch status {
	case "prospect", "active", "former", "blocked":
		return true
	default:
		return false
	}
}
