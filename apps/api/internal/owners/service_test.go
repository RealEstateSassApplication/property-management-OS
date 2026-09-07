package owners

import (
	"context"
	"testing"
)

type fakeRepository struct {
	ownerInput    CreateOwnerInput
	interestInput CreateInterestInput
	ownerCalls    int
	interestCalls int
}

func (f *fakeRepository) List(context.Context, string) ([]Owner, error) { return nil, nil }
func (f *fakeRepository) Get(context.Context, string, string) (Owner, error) { return Owner{}, nil }
func (f *fakeRepository) ListInterests(context.Context, string) ([]OwnershipInterest, error) { return nil, nil }
func (f *fakeRepository) Create(_ context.Context, organizationID string, input CreateOwnerInput) (Owner, error) {
	f.ownerCalls++
	f.ownerInput = input
	return Owner{OrganizationID: organizationID, LegalName: input.LegalName, OwnerType: input.OwnerType, Email: input.Email, Status: input.Status}, nil
}
func (f *fakeRepository) CreateInterest(_ context.Context, organizationID string, input CreateInterestInput) (OwnershipInterest, error) {
	f.interestCalls++
	f.interestInput = input
	return OwnershipInterest{OrganizationID: organizationID, OwnerID: input.OwnerID, PropertyID: input.PropertyID, OwnershipBps: input.OwnershipBps, EffectiveFrom: input.EffectiveFrom}, nil
}

func TestCreateNormalizesOwner(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	owner, err := service.Create(context.Background(), "org", CreateOwnerInput{
		LegalName: "  Maya Holdings  ",
		OwnerType: " COMPANY ",
		Email:     " OWNER@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner.LegalName != "Maya Holdings" || repo.ownerInput.Email != "owner@example.com" || repo.ownerInput.Status != "active" {
		t.Fatalf("unexpected normalized owner: %#v", repo.ownerInput)
	}
}

func TestCreateInterestRejectsInvalidPercentage(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.CreateInterest(context.Background(), "org", CreateInterestInput{
		OwnerID: "owner", PropertyID: "property", OwnershipBps: 10001, EffectiveFrom: "2026-09-01",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.interestCalls != 0 {
		t.Fatal("repository should not be called")
	}
}
