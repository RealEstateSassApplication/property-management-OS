package tenants

import (
	"context"
	"testing"
)

type fakeRepository struct {
	createInput CreateInput
	createCalls int
}

func (f *fakeRepository) List(context.Context, string) ([]Tenant, error)      { return nil, nil }
func (f *fakeRepository) Get(context.Context, string, string) (Tenant, error) { return Tenant{}, nil }
func (f *fakeRepository) Update(context.Context, string, string, UpdateInput) (Tenant, error) {
	return Tenant{}, nil
}
func (f *fakeRepository) Create(_ context.Context, organizationID string, input CreateInput) (Tenant, error) {
	f.createCalls++
	f.createInput = input
	return Tenant{OrganizationID: organizationID, LegalName: input.LegalName, Email: input.Email, Status: input.Status}, nil
}

func TestCreateNormalizesTenant(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	tenant, err := service.Create(context.Background(), "org", CreateInput{
		LegalName: "  Maya Silva  ",
		Email:     " MAYA@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createInput.LegalName != "Maya Silva" || repo.createInput.Email != "maya@example.com" || repo.createInput.Status != "prospect" {
		t.Fatalf("unexpected normalized input: %#v", repo.createInput)
	}
	if tenant.LegalName != "Maya Silva" {
		t.Fatalf("expected normalized tenant")
	}
}

func TestCreateRejectsInvalidEmail(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	if _, err := service.Create(context.Background(), "org", CreateInput{LegalName: "Maya Silva", Email: "not-an-email"}); err == nil {
		t.Fatal("expected validation error")
	}
	if repo.createCalls != 0 {
		t.Fatal("repository should not be called")
	}
}
