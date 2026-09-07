package rent

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	ListObligations(ctx context.Context, organizationID string) ([]Obligation, error)
	CreateObligation(ctx context.Context, organizationID, leaseID string, periodStart time.Time) (Obligation, error)
	ListPayments(ctx context.Context, organizationID string) ([]Payment, error)
	CreatePayment(ctx context.Context, organizationID string, input CreatePaymentInput) (Payment, error)
	CreateAllocation(ctx context.Context, organizationID string, input CreateAllocationInput) (Allocation, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListObligations(ctx context.Context, organizationID string) ([]Obligation, error) {
	return s.repository.ListObligations(ctx, organizationID)
}

func (s *Service) CreateObligation(ctx context.Context, organizationID string, input CreateObligationInput) (Obligation, error) {
	input.LeaseID = strings.TrimSpace(input.LeaseID)
	input.Period = strings.TrimSpace(input.Period)
	if input.LeaseID == "" {
		return Obligation{}, fmt.Errorf("leaseId is required")
	}
	periodStart, err := time.Parse("2006-01", input.Period)
	if err != nil {
		return Obligation{}, fmt.Errorf("period must be YYYY-MM")
	}
	return s.repository.CreateObligation(ctx, organizationID, input.LeaseID, periodStart.UTC())
}

func (s *Service) ListPayments(ctx context.Context, organizationID string) ([]Payment, error) {
	return s.repository.ListPayments(ctx, organizationID)
}

func (s *Service) CreatePayment(ctx context.Context, organizationID string, input CreatePaymentInput) (Payment, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.ReceivedAt = strings.TrimSpace(input.ReceivedAt)
	input.Method = strings.ToLower(strings.TrimSpace(input.Method))
	input.ReferenceCode = strings.TrimSpace(input.ReferenceCode)
	if input.TenantID == "" {
		return Payment{}, fmt.Errorf("tenantId is required")
	}
	if input.AmountMinor <= 0 {
		return Payment{}, fmt.Errorf("amountMinor must be greater than zero")
	}
	if len(input.Currency) != 3 {
		return Payment{}, fmt.Errorf("currency must be a 3-letter code")
	}
	if _, err := time.Parse("2006-01-02", input.ReceivedAt); err != nil {
		return Payment{}, fmt.Errorf("receivedAt must be YYYY-MM-DD")
	}
	switch input.Method {
	case "cash", "bank_transfer", "card", "online", "other":
	default:
		return Payment{}, fmt.Errorf("method is invalid")
	}
	return s.repository.CreatePayment(ctx, organizationID, input)
}

func (s *Service) CreateAllocation(ctx context.Context, organizationID string, input CreateAllocationInput) (Allocation, error) {
	input.PaymentID = strings.TrimSpace(input.PaymentID)
	input.ObligationID = strings.TrimSpace(input.ObligationID)
	if input.PaymentID == "" || input.ObligationID == "" {
		return Allocation{}, fmt.Errorf("paymentId and obligationId are required")
	}
	if input.AmountMinor <= 0 {
		return Allocation{}, fmt.Errorf("amountMinor must be greater than zero")
	}
	return s.repository.CreateAllocation(ctx, organizationID, input)
}
