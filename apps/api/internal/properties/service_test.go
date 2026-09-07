package properties

import (
	"context"
	"testing"
)

type fakeRepository struct {
	createInput CreateInput
	createCalls int
}

func (f *fakeRepository) List(context.Context, string) ([]Property, error) {
	return nil, nil
}

func (f *fakeRepository) Get(context.Context, string, string) (Property, error) {
	return Property{}, nil
}

func (f *fakeRepository) Update(context.Context, string, string, UpdateInput) (Property, error) {
	return Property{}, nil
}

func (f *fakeRepository) Create(_ context.Context, organizationID string, input CreateInput) (Property, error) {
	f.createCalls++
	f.createInput = input
	return Property{OrganizationID: organizationID, Name: input.Name, CountryCode: input.CountryCode}, nil
}

func TestCreateValidatesAndNormalizesProperty(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	property, err := service.Create(context.Background(), "org-1", CreateInput{
		Name:         "  Lake House  ",
		PropertyType: "HOUSE",
		AddressLine1: "  10 Lake Road ",
		City:         " Colombo ",
		CountryCode:  "lk",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repository create once, got %d", repo.createCalls)
	}
	if repo.createInput.PropertyType != "house" || repo.createInput.CountryCode != "LK" {
		t.Fatalf("expected normalized input, got %#v", repo.createInput)
	}
	if property.Name != "Lake House" {
		t.Fatalf("expected trimmed name, got %q", property.Name)
	}
}

func TestCreateRejectsInvalidProperty(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	_, err := service.Create(context.Background(), "org-1", CreateInput{
		Name:         "Bad Property",
		PropertyType: "castle",
		AddressLine1: "Road",
		City:         "Colombo",
		CountryCode:  "LK",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if repo.createCalls != 0 {
		t.Fatalf("repository should not be called for invalid input")
	}
}
