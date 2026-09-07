package tenancies

import (
	"context"
	"testing"
)

type fakeRepository struct {
	createInput CreateInput
	createCalls int
}

func (f *fakeRepository) List(context.Context, string) ([]Tenancy, error) { return nil, nil }
func (f *fakeRepository) Get(context.Context, string, string) (Tenancy, error) { return Tenancy{}, nil }
func (f *fakeRepository) Update(context.Context, string, string, UpdateInput) (Tenancy, error) {
	return Tenancy{}, nil
}
func (f *fakeRepository) Create(_ context.Context, organizationID string, input CreateInput) (Tenancy, error) {
	f.createCalls++
	f.createInput = input
	return Tenancy{OrganizationID: organizationID, UnitID: input.UnitID, PrimaryTenantID: input.PrimaryTenantID, Status: input.Status}, nil
}

func TestCreateNormalizesAndDeduplicatesOccupants(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	_, err := service.Create(context.Background(), "org", CreateInput{
		UnitID:            " unit-1 ",
		PrimaryTenantID:   " tenant-1 ",
		OccupantTenantIDs: []string{"tenant-1", " tenant-2 ", "tenant-2"},
		StartDate:         "2026-09-15",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createInput.Status != "upcoming" || len(repo.createInput.OccupantTenantIDs) != 1 || repo.createInput.OccupantTenantIDs[0] != "tenant-2" {
		t.Fatalf("unexpected normalized input: %#v", repo.createInput)
	}
}

func TestCreateRejectsEndBeforeStart(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.Create(context.Background(), "org", CreateInput{
		UnitID:          "unit-1",
		PrimaryTenantID: "tenant-1",
		StartDate:       "2026-09-15",
		EndDate:         "2026-09-14",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.createCalls != 0 {
		t.Fatal("repository should not be called")
	}
}
