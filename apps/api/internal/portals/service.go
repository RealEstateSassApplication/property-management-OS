package portals

import (
	"context"
	"fmt"
	"strings"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
)

type Repository interface {
	ListOwnerProfiles(ctx context.Context, organizationID, userID string) ([]OwnerProfile, error)
	ListOwnerPropertyBases(ctx context.Context, organizationID, userID string) ([]OwnerPropertyBase, error)
	ListOwnerReceivables(ctx context.Context, organizationID, userID string) ([]OwnerReceivableRow, error)
	ListTenantProfiles(ctx context.Context, organizationID, userID string) ([]TenantProfile, error)
	ListTenantOccupancies(ctx context.Context, organizationID, userID string) ([]TenantOccupancy, error)
	ListTenantRent(ctx context.Context, organizationID, userID string) ([]TenantRentItem, error)
	ListTenantMaintenance(ctx context.Context, organizationID, userID string) ([]TenantMaintenanceItem, error)
	GetTenantMaintenanceContext(ctx context.Context, organizationID, userID, tenancyID string) (TenantMaintenanceContext, error)
}

type MaintenanceCreator interface {
	CreateRequest(ctx context.Context, organizationID, actorUserID string, input maintenance.CreateRequestInput) (maintenance.Request, error)
}

type Service struct {
	repository  Repository
	maintenance MaintenanceCreator
}

func NewService(repository Repository, maintenanceCreator MaintenanceCreator) *Service {
	return &Service{repository: repository, maintenance: maintenanceCreator}
}

func (s *Service) OwnerSummary(ctx context.Context, organizationID, userID string) (OwnerSummary, error) {
	owners, err := s.repository.ListOwnerProfiles(ctx, organizationID, userID)
	if err != nil {
		return OwnerSummary{}, err
	}
	if len(owners) == 0 {
		return OwnerSummary{}, ErrOwnerLinkNotFound
	}
	bases, err := s.repository.ListOwnerPropertyBases(ctx, organizationID, userID)
	if err != nil {
		return OwnerSummary{}, err
	}
	receivables, err := s.repository.ListOwnerReceivables(ctx, organizationID, userID)
	if err != nil {
		return OwnerSummary{}, err
	}
	byProperty := make(map[string]map[string]int64)
	for _, row := range receivables {
		if _, ok := byProperty[row.PropertyID]; !ok {
			byProperty[row.PropertyID] = make(map[string]int64)
		}
		byProperty[row.PropertyID][row.Currency] += row.OutstandingMinor
	}
	properties := make([]OwnerProperty, 0, len(bases))
	for _, base := range bases {
		outstanding := byProperty[base.PropertyID]
		if outstanding == nil {
			outstanding = map[string]int64{}
		}
		properties = append(properties, OwnerProperty{
			OwnerID: base.OwnerID, PropertyID: base.PropertyID, ReferenceCode: base.ReferenceCode,
			Name: base.Name, City: base.City, OwnershipBPS: base.OwnershipBPS,
			UnitCount: base.UnitCount, OccupiedUnits: base.OccupiedUnits,
			OpenMaintenance: base.OpenMaintenance, OutstandingByCurrency: outstanding,
		})
	}
	return OwnerSummary{Owners: owners, Properties: properties}, nil
}

func (s *Service) TenantSummary(ctx context.Context, organizationID, userID string) (TenantSummary, error) {
	tenants, err := s.repository.ListTenantProfiles(ctx, organizationID, userID)
	if err != nil {
		return TenantSummary{}, err
	}
	if len(tenants) == 0 {
		return TenantSummary{}, ErrTenantLinkNotFound
	}
	occupancies, err := s.repository.ListTenantOccupancies(ctx, organizationID, userID)
	if err != nil {
		return TenantSummary{}, err
	}
	rentItems, err := s.repository.ListTenantRent(ctx, organizationID, userID)
	if err != nil {
		return TenantSummary{}, err
	}
	maintenanceItems, err := s.repository.ListTenantMaintenance(ctx, organizationID, userID)
	if err != nil {
		return TenantSummary{}, err
	}
	return TenantSummary{Tenants: tenants, Occupancies: occupancies, Rent: rentItems, Maintenance: maintenanceItems}, nil
}

func (s *Service) CreateTenantMaintenance(ctx context.Context, organizationID, userID string, input CreateTenantMaintenanceInput) (maintenance.Request, error) {
	input.TenancyID = strings.TrimSpace(input.TenancyID)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	input.Priority = strings.ToLower(strings.TrimSpace(input.Priority))
	if input.TenancyID == "" || input.Title == "" || input.Description == "" {
		return maintenance.Request{}, fmt.Errorf("tenancyId, title and description are required")
	}
	context, err := s.repository.GetTenantMaintenanceContext(ctx, organizationID, userID, input.TenancyID)
	if err != nil {
		return maintenance.Request{}, err
	}
	return s.maintenance.CreateRequest(ctx, organizationID, userID, maintenance.CreateRequestInput{
		PropertyID:  context.PropertyID,
		UnitID:      context.UnitID,
		TenantID:    context.TenantID,
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		Priority:    input.Priority,
	})
}
