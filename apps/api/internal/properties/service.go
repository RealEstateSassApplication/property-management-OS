package properties

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("property not found")

type Repository interface {
	List(context.Context, string) ([]Property, error)
	Get(context.Context, string, string) (Property, error)
	Create(context.Context, string, CreateInput) (Property, error)
	Update(context.Context, string, string, UpdateInput) (Property, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, organizationID string) ([]Property, error) {
	return s.repo.List(ctx, organizationID)
}

func (s *Service) Get(ctx context.Context, organizationID, id string) (Property, error) {
	return s.repo.Get(ctx, organizationID, id)
}

func (s *Service) Create(ctx context.Context, organizationID string, input CreateInput) (Property, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.PropertyType = strings.ToLower(strings.TrimSpace(input.PropertyType))
	input.AddressLine1 = strings.TrimSpace(input.AddressLine1)
	input.City = strings.TrimSpace(input.City)
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))

	if err := validateCreate(input); err != nil {
		return Property{}, err
	}
	return s.repo.Create(ctx, organizationID, input)
}

func (s *Service) Update(ctx context.Context, organizationID, id string, input UpdateInput) (Property, error) {
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		input.Name = &value
	}
	if input.PropertyType != nil {
		value := strings.ToLower(strings.TrimSpace(*input.PropertyType))
		input.PropertyType = &value
	}
	if input.CountryCode != nil {
		value := strings.ToUpper(strings.TrimSpace(*input.CountryCode))
		input.CountryCode = &value
	}
	if err := validateUpdate(input); err != nil {
		return Property{}, err
	}
	return s.repo.Update(ctx, organizationID, id, input)
}

func validateCreate(input CreateInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !validPropertyType(input.PropertyType) {
		return fmt.Errorf("invalid propertyType")
	}
	if input.AddressLine1 == "" || input.City == "" {
		return fmt.Errorf("addressLine1 and city are required")
	}
	if len(input.CountryCode) != 2 {
		return fmt.Errorf("countryCode must be a 2-letter ISO code")
	}
	return nil
}

func validateUpdate(input UpdateInput) error {
	if input.Name != nil && *input.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if input.PropertyType != nil && !validPropertyType(*input.PropertyType) {
		return fmt.Errorf("invalid propertyType")
	}
	if input.CountryCode != nil && len(*input.CountryCode) != 2 {
		return fmt.Errorf("countryCode must be a 2-letter ISO code")
	}
	if input.Status != nil && *input.Status != "active" && *input.Status != "inactive" && *input.Status != "archived" {
		return fmt.Errorf("invalid status")
	}
	return nil
}

func validPropertyType(value string) bool {
	switch value {
	case "house", "apartment", "building", "commercial", "land", "other":
		return true
	default:
		return false
	}
}
