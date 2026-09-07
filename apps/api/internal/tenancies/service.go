package tenancies

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound       = errors.New("tenancy not found")
	ErrUnitConflict   = errors.New("unit already has an active tenancy")
	ErrTenantInvalid  = errors.New("tenant is not eligible for tenancy")
	ErrUnitNotFound   = errors.New("unit not found")
	ErrTenantNotFound = errors.New("tenant not found")
)

type Repository interface {
	List(ctx context.Context, organizationID string) ([]Tenancy, error)
	Get(ctx context.Context, organizationID, tenancyID string) (Tenancy, error)
	Create(ctx context.Context, organizationID string, input CreateInput) (Tenancy, error)
	Update(ctx context.Context, organizationID, tenancyID string, input UpdateInput) (Tenancy, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, organizationID string) ([]Tenancy, error) {
	return s.repository.List(ctx, organizationID)
}

func (s *Service) Get(ctx context.Context, organizationID, tenancyID string) (Tenancy, error) {
	return s.repository.Get(ctx, organizationID, tenancyID)
}

func (s *Service) Create(ctx context.Context, organizationID string, input CreateInput) (Tenancy, error) {
	normalized, err := normalizeCreate(input)
	if err != nil {
		return Tenancy{}, err
	}
	return s.repository.Create(ctx, organizationID, normalized)
}

func (s *Service) Update(ctx context.Context, organizationID, tenancyID string, input UpdateInput) (Tenancy, error) {
	if input.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*input.Status))
		if !validStatus(status) {
			return Tenancy{}, fmt.Errorf("status must be one of upcoming, active, ended, cancelled")
		}
		input.Status = &status
	}
	if input.EndDate != nil {
		value := strings.TrimSpace(*input.EndDate)
		if value != "" {
			if _, err := time.Parse(time.DateOnly, value); err != nil {
				return Tenancy{}, fmt.Errorf("endDate must use YYYY-MM-DD")
			}
		}
		input.EndDate = &value
	}
	return s.repository.Update(ctx, organizationID, tenancyID, input)
}

func normalizeCreate(input CreateInput) (CreateInput, error) {
	input.UnitID = strings.TrimSpace(input.UnitID)
	input.PrimaryTenantID = strings.TrimSpace(input.PrimaryTenantID)
	input.StartDate = strings.TrimSpace(input.StartDate)
	input.EndDate = strings.TrimSpace(input.EndDate)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "upcoming"
	}
	if input.UnitID == "" || input.PrimaryTenantID == "" {
		return CreateInput{}, fmt.Errorf("unitId and primaryTenantId are required")
	}
	start, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		return CreateInput{}, fmt.Errorf("startDate must use YYYY-MM-DD")
	}
	if input.EndDate != "" {
		end, err := time.Parse(time.DateOnly, input.EndDate)
		if err != nil {
			return CreateInput{}, fmt.Errorf("endDate must use YYYY-MM-DD")
		}
		if end.Before(start) {
			return CreateInput{}, fmt.Errorf("endDate cannot be before startDate")
		}
	}
	if !validStatus(input.Status) {
		return CreateInput{}, fmt.Errorf("status must be one of upcoming, active, ended, cancelled")
	}

	seen := map[string]struct{}{input.PrimaryTenantID: {}}
	occupants := make([]string, 0, len(input.OccupantTenantIDs))
	for _, tenantID := range input.OccupantTenantIDs {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID == "" {
			continue
		}
		if _, exists := seen[tenantID]; exists {
			continue
		}
		seen[tenantID] = struct{}{}
		occupants = append(occupants, tenantID)
	}
	input.OccupantTenantIDs = occupants
	return input, nil
}

func validStatus(status string) bool {
	switch status {
	case "upcoming", "active", "ended", "cancelled":
		return true
	default:
		return false
	}
}
