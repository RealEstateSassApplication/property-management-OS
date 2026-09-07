package inspections

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	created CreateInspectionInput
	item    CreateItemInput
}

func (f *fakeRepository) List(context.Context, string) ([]Inspection, error) { return nil, nil }
func (f *fakeRepository) Get(context.Context, string, string) (Inspection, error) {
	return Inspection{}, nil
}
func (f *fakeRepository) ListItems(context.Context, string, string) ([]Item, error) { return nil, nil }
func (f *fakeRepository) Create(_ context.Context, _, _ string, input CreateInspectionInput) (Inspection, error) {
	f.created = input
	return Inspection{InspectionType: input.InspectionType}, nil
}
func (f *fakeRepository) CreateItem(_ context.Context, _ string, input CreateItemInput) (Item, error) {
	f.item = input
	return Item{Condition: input.Condition}, nil
}
func (f *fakeRepository) Complete(context.Context, string, string, CompleteInspectionInput) (Inspection, error) {
	return Inspection{}, nil
}
func (f *fakeRepository) Acknowledge(context.Context, string, string, AcknowledgeInspectionInput) (Inspection, error) {
	return Inspection{}, nil
}

func TestCreateNormalizesInspectionType(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	item, err := service.Create(context.Background(), "org", "user", CreateInspectionInput{TenancyID: " tenancy ", InspectionType: " MOVE_IN ", ScheduledFor: "2026-09-08T09:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	if item.InspectionType != "move_in" || repo.created.TenancyID != "tenancy" {
		t.Fatalf("unexpected normalized input: %+v", repo.created)
	}
}

func TestCreateRejectsUnsupportedType(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Create(context.Background(), "org", "user", CreateInspectionInput{TenancyID: "t", InspectionType: "fire"})
	if !errors.Is(err, ErrInvalidInspectionType) {
		t.Fatalf("expected ErrInvalidInspectionType, got %v", err)
	}
}

func TestCreateItemValidatesCondition(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.CreateItem(context.Background(), "org", CreateItemInput{InspectionID: "i", Area: "Kitchen", ItemName: "Sink", Condition: "excellent"})
	if !errors.Is(err, ErrInvalidInspectionItem) {
		t.Fatalf("expected ErrInvalidInspectionItem, got %v", err)
	}
}
