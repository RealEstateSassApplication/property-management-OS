package leases

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	lease       Lease
	createInput CreateInput
	createCalls int
	updateInput UpdateInput
}

func (f *fakeRepository) List(context.Context, string) ([]Lease, error)      { return nil, nil }
func (f *fakeRepository) Get(context.Context, string, string) (Lease, error) { return f.lease, nil }
func (f *fakeRepository) Create(_ context.Context, organizationID string, input CreateInput) (Lease, error) {
	f.createCalls++
	f.createInput = input
	return Lease{OrganizationID: organizationID, TenancyID: input.TenancyID, ReferenceCode: input.ReferenceCode, Currency: input.Currency, Status: input.Status}, nil
}
func (f *fakeRepository) Update(_ context.Context, _ string, _ string, input UpdateInput) (Lease, error) {
	f.updateInput = input
	return Lease{Status: *input.Status}, nil
}

func TestCreateNormalizesLease(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.Create(context.Background(), "org", CreateInput{
		TenancyID:          " tenancy-1 ",
		ReferenceCode:      " LEASE-001 ",
		StartDate:          "2026-10-01",
		EndDate:            "2027-09-30",
		RentAmountMinor:    15000000,
		DepositAmountMinor: 30000000,
		Currency:           " lkr ",
		DueDay:             1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createInput.Currency != "LKR" || repo.createInput.Status != "draft" || repo.createInput.ReferenceCode != "LEASE-001" {
		t.Fatalf("unexpected normalized input: %#v", repo.createInput)
	}
}

func TestCreateRejectsFloatingStyleInvalidMoney(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.Create(context.Background(), "org", CreateInput{
		TenancyID:       "tenancy-1",
		ReferenceCode:   "LEASE-001",
		StartDate:       "2026-10-01",
		EndDate:         "2027-09-30",
		RentAmountMinor: 0,
		Currency:        "LKR",
		DueDay:          1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.createCalls != 0 {
		t.Fatal("repository should not be called")
	}
}

func TestUpdateRejectsInvalidTransition(t *testing.T) {
	repo := &fakeRepository{lease: Lease{Status: "expired"}}
	service := NewService(repo)
	status := "active"
	_, err := service.Update(context.Background(), "org", "lease", UpdateInput{Status: &status})
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}
