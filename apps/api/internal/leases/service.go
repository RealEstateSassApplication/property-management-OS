package leases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound            = errors.New("lease not found")
	ErrTenancyNotFound     = errors.New("tenancy not found")
	ErrActiveLeaseConflict = errors.New("tenancy already has an active lease")
	ErrInvalidTransition   = errors.New("invalid lease status transition")
)

type Repository interface {
	List(ctx context.Context, organizationID string) ([]Lease, error)
	Get(ctx context.Context, organizationID, leaseID string) (Lease, error)
	Create(ctx context.Context, organizationID string, input CreateInput) (Lease, error)
	Update(ctx context.Context, organizationID, leaseID string, input UpdateInput) (Lease, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, organizationID string) ([]Lease, error) {
	return s.repository.List(ctx, organizationID)
}

func (s *Service) Get(ctx context.Context, organizationID, leaseID string) (Lease, error) {
	return s.repository.Get(ctx, organizationID, leaseID)
}

func (s *Service) Create(ctx context.Context, organizationID string, input CreateInput) (Lease, error) {
	normalized, err := normalizeCreate(input)
	if err != nil {
		return Lease{}, err
	}
	return s.repository.Create(ctx, organizationID, normalized)
}

func (s *Service) Update(ctx context.Context, organizationID, leaseID string, input UpdateInput) (Lease, error) {
	if input.Status == nil {
		return s.repository.Get(ctx, organizationID, leaseID)
	}
	status := strings.ToLower(strings.TrimSpace(*input.Status))
	if !validStatus(status) {
		return Lease{}, fmt.Errorf("status must be one of draft, active, expired, terminated, cancelled")
	}
	current, err := s.repository.Get(ctx, organizationID, leaseID)
	if err != nil {
		return Lease{}, err
	}
	if !validTransition(current.Status, status) {
		return Lease{}, ErrInvalidTransition
	}
	input.Status = &status
	return s.repository.Update(ctx, organizationID, leaseID, input)
}

func normalizeCreate(input CreateInput) (CreateInput, error) {
	input.TenancyID = strings.TrimSpace(input.TenancyID)
	input.ReferenceCode = strings.TrimSpace(input.ReferenceCode)
	input.StartDate = strings.TrimSpace(input.StartDate)
	input.EndDate = strings.TrimSpace(input.EndDate)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "draft"
	}
	if input.TenancyID == "" || input.ReferenceCode == "" {
		return CreateInput{}, fmt.Errorf("tenancyId and referenceCode are required")
	}
	start, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		return CreateInput{}, fmt.Errorf("startDate must use YYYY-MM-DD")
	}
	end, err := time.Parse(time.DateOnly, input.EndDate)
	if err != nil {
		return CreateInput{}, fmt.Errorf("endDate must use YYYY-MM-DD")
	}
	if end.Before(start) {
		return CreateInput{}, fmt.Errorf("endDate cannot be before startDate")
	}
	if input.RentAmountMinor <= 0 {
		return CreateInput{}, fmt.Errorf("rentAmountMinor must be greater than zero")
	}
	if input.DepositAmountMinor < 0 {
		return CreateInput{}, fmt.Errorf("depositAmountMinor cannot be negative")
	}
	if len(input.Currency) != 3 {
		return CreateInput{}, fmt.Errorf("currency must be a 3-letter ISO code")
	}
	if input.DueDay < 1 || input.DueDay > 31 {
		return CreateInput{}, fmt.Errorf("dueDay must be between 1 and 31")
	}
	if !validStatus(input.Status) {
		return CreateInput{}, fmt.Errorf("status must be one of draft, active, expired, terminated, cancelled")
	}
	return input, nil
}

func validStatus(status string) bool {
	switch status {
	case "draft", "active", "expired", "terminated", "cancelled":
		return true
	default:
		return false
	}
}

func validTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "draft":
		return to == "active" || to == "cancelled"
	case "active":
		return to == "expired" || to == "terminated"
	default:
		return false
	}
}
