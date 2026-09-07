package units

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("unit not found")

type Repository interface {
	ListByProperty(context.Context, string, string) ([]Unit, error)
	Get(context.Context, string, string) (Unit, error)
	Create(context.Context, string, string, CreateInput) (Unit, error)
	Update(context.Context, string, string, UpdateInput) (Unit, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByProperty(ctx context.Context, organizationID, propertyID string) ([]Unit, error) {
	return s.repo.ListByProperty(ctx, organizationID, propertyID)
}

func (s *Service) Get(ctx context.Context, organizationID, id string) (Unit, error) {
	return s.repo.Get(ctx, organizationID, id)
}

func (s *Service) Create(ctx context.Context, organizationID, propertyID string, input CreateInput) (Unit, error) {
	input.ReferenceCode = strings.TrimSpace(input.ReferenceCode)
	input.Label = strings.TrimSpace(input.Label)
	if input.FloorAreaUnit != nil {
		value := strings.ToLower(strings.TrimSpace(*input.FloorAreaUnit))
		input.FloorAreaUnit = &value
	}
	if err := validateCreate(input); err != nil {
		return Unit{}, err
	}
	return s.repo.Create(ctx, organizationID, propertyID, input)
}

func (s *Service) Update(ctx context.Context, organizationID, id string, input UpdateInput) (Unit, error) {
	if input.ReferenceCode != nil {
		value := strings.TrimSpace(*input.ReferenceCode)
		input.ReferenceCode = &value
	}
	if input.Label != nil {
		value := strings.TrimSpace(*input.Label)
		input.Label = &value
	}
	if input.FloorAreaUnit != nil {
		value := strings.ToLower(strings.TrimSpace(*input.FloorAreaUnit))
		input.FloorAreaUnit = &value
	}
	if input.OccupancyStatus != nil {
		value := strings.ToLower(strings.TrimSpace(*input.OccupancyStatus))
		input.OccupancyStatus = &value
	}
	if err := validateUpdate(input); err != nil {
		return Unit{}, err
	}
	return s.repo.Update(ctx, organizationID, id, input)
}

func validateCreate(input CreateInput) error {
	if input.ReferenceCode == "" || input.Label == "" {
		return fmt.Errorf("referenceCode and label are required")
	}
	return validateMeasurements(input.Bedrooms, input.Bathrooms, input.FloorArea, input.FloorAreaUnit)
}

func validateUpdate(input UpdateInput) error {
	if input.ReferenceCode != nil && *input.ReferenceCode == "" {
		return fmt.Errorf("referenceCode cannot be empty")
	}
	if input.Label != nil && *input.Label == "" {
		return fmt.Errorf("label cannot be empty")
	}
	if input.OccupancyStatus != nil {
		switch *input.OccupancyStatus {
		case "vacant", "occupied", "reserved", "unavailable":
		default:
			return fmt.Errorf("invalid occupancyStatus")
		}
	}
	return validateMeasurements(input.Bedrooms, input.Bathrooms, input.FloorArea, input.FloorAreaUnit)
}

func validateMeasurements(bedrooms, bathrooms, floorArea *float64, floorAreaUnit *string) error {
	for _, value := range []*float64{bedrooms, bathrooms, floorArea} {
		if value != nil && *value < 0 {
			return fmt.Errorf("measurements cannot be negative")
		}
	}
	if floorAreaUnit != nil && *floorAreaUnit != "sqft" && *floorAreaUnit != "sqm" {
		return fmt.Errorf("floorAreaUnit must be sqft or sqm")
	}
	return nil
}
