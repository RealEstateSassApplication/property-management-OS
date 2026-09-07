package rent

import (
	"context"
	"testing"
	"time"
)

type fakeRepository struct {
	obligationLeaseID string
	obligationPeriod  time.Time
	paymentInput      CreatePaymentInput
	allocationInput   CreateAllocationInput
	obligationCalls   int
	paymentCalls      int
	allocationCalls   int
}

func (f *fakeRepository) ListObligations(context.Context, string) ([]Obligation, error) {
	return nil, nil
}
func (f *fakeRepository) ListPayments(context.Context, string) ([]Payment, error) { return nil, nil }
func (f *fakeRepository) CreateObligation(_ context.Context, organizationID, leaseID string, periodStart time.Time) (Obligation, error) {
	f.obligationCalls++
	f.obligationLeaseID = leaseID
	f.obligationPeriod = periodStart
	return Obligation{OrganizationID: organizationID, LeaseID: leaseID, Period: periodStart.Format("2006-01")}, nil
}
func (f *fakeRepository) CreatePayment(_ context.Context, organizationID string, input CreatePaymentInput) (Payment, error) {
	f.paymentCalls++
	f.paymentInput = input
	return Payment{OrganizationID: organizationID, TenantID: input.TenantID, AmountMinor: input.AmountMinor, Currency: input.Currency}, nil
}
func (f *fakeRepository) CreateAllocation(_ context.Context, organizationID string, input CreateAllocationInput) (Allocation, error) {
	f.allocationCalls++
	f.allocationInput = input
	return Allocation{OrganizationID: organizationID, PaymentID: input.PaymentID, ObligationID: input.ObligationID, AmountMinor: input.AmountMinor}, nil
}

func TestCreateObligationNormalizesMonth(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreateObligation(context.Background(), "org", CreateObligationInput{LeaseID: " lease-1 ", Period: "2026-09"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.obligationLeaseID != "lease-1" || repo.obligationPeriod.Format("2006-01-02") != "2026-09-01" {
		t.Fatalf("unexpected obligation normalization: %q %s", repo.obligationLeaseID, repo.obligationPeriod)
	}
}

func TestCreatePaymentNormalizesCurrencyAndMethod(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreatePayment(context.Background(), "org", CreatePaymentInput{
		TenantID: " tenant-1 ", AmountMinor: 15000000, Currency: " lkr ", ReceivedAt: "2026-09-07", Method: " BANK_TRANSFER ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.paymentInput.Currency != "LKR" || repo.paymentInput.Method != "bank_transfer" || repo.paymentInput.TenantID != "tenant-1" {
		t.Fatalf("unexpected payment normalization: %#v", repo.paymentInput)
	}
}

func TestCreateAllocationRejectsNonPositiveAmount(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreateAllocation(context.Background(), "org", CreateAllocationInput{PaymentID: "payment", ObligationID: "obligation", AmountMinor: 0})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.allocationCalls != 0 {
		t.Fatal("repository should not be called")
	}
}
