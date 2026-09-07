package units

import (
	"context"
	"testing"
)

type fakeRepository struct {
	createInput CreateInput
	createCalls int
}

func (f *fakeRepository) ListByProperty(context.Context, string, string) ([]Unit, error) { return nil, nil }
func (f *fakeRepository) Get(context.Context, string, string) (Unit, error) { return Unit{}, nil }
func (f *fakeRepository) Update(context.Context, string, string, UpdateInput) (Unit, error) {
	return Unit{}, nil
}
func (f *fakeRepository) Create(_ context.Context, organizationID, propertyID string, input CreateInput) (Unit, error) {
	f.createCalls++
	f.createInput = input
	return Unit{OrganizationID: organizationID, PropertyID: propertyID, ReferenceCode: input.ReferenceCode, Label: input.Label}, nil
}

func TestCreateNormalizesUnit(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	areaUnit := " SQM "

	unit, err := service.Create(context.Background(), "org-1", "property-1", CreateInput{
		ReferenceCode: "  A-101 ",
		Label:         " Unit 101 ",
		FloorAreaUnit: &areaUnit,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repository create once")
	}
	if repo.createInput.ReferenceCode != "A-101" || *repo.createInput.FloorAreaUnit != "sqm" {
		t.Fatalf("expected normalized input, got %#v", repo.createInput)
	}
	if unit.Label != "Unit 101" {
		t.Fatalf("expected trimmed label, got %q", unit.Label)
	}
}

func TestCreateRejectsNegativeMeasurements(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	bedrooms := -1.0

	_, err := service.Create(context.Background(), "org-1", "property-1", CreateInput{
		ReferenceCode: "A-101",
		Label:         "Unit 101",
		Bedrooms:      &bedrooms,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.createCalls != 0 {
		t.Fatal("repository should not be called for invalid input")
	}
}
