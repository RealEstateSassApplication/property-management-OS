package portals

import (
	"context"
	"errors"
	"testing"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
)

type fakeRepository struct {
	owners       []OwnerProfile
	ownerBases   []OwnerPropertyBase
	receivables  []OwnerReceivableRow
	tenants      []TenantProfile
	occupancies  []TenantOccupancy
	rent         []TenantRentItem
	maintenance  []TenantMaintenanceItem
	context      TenantMaintenanceContext
	contextError error
}

func (f fakeRepository) ListOwnerProfiles(context.Context, string, string) ([]OwnerProfile, error) {
	return f.owners, nil
}
func (f fakeRepository) ListOwnerPropertyBases(context.Context, string, string) ([]OwnerPropertyBase, error) {
	return f.ownerBases, nil
}
func (f fakeRepository) ListOwnerReceivables(context.Context, string, string) ([]OwnerReceivableRow, error) {
	return f.receivables, nil
}
func (f fakeRepository) ListTenantProfiles(context.Context, string, string) ([]TenantProfile, error) {
	return f.tenants, nil
}
func (f fakeRepository) ListTenantOccupancies(context.Context, string, string) ([]TenantOccupancy, error) {
	return f.occupancies, nil
}
func (f fakeRepository) ListTenantRent(context.Context, string, string) ([]TenantRentItem, error) {
	return f.rent, nil
}
func (f fakeRepository) ListTenantMaintenance(context.Context, string, string) ([]TenantMaintenanceItem, error) {
	return f.maintenance, nil
}
func (f fakeRepository) GetTenantMaintenanceContext(context.Context, string, string, string) (TenantMaintenanceContext, error) {
	return f.context, f.contextError
}

type fakeMaintenanceCreator struct {
	input maintenance.CreateRequestInput
}

func (f *fakeMaintenanceCreator) CreateRequest(_ context.Context, _, _ string, input maintenance.CreateRequestInput) (maintenance.Request, error) {
	f.input = input
	return maintenance.Request{ID: "request-1", PropertyID: input.PropertyID, UnitID: input.UnitID, TenantID: input.TenantID, Title: input.Title, Status: "open"}, nil
}

func TestOwnerSummaryRequiresExplicitOwnerLink(t *testing.T) {
	service := NewService(fakeRepository{}, &fakeMaintenanceCreator{})
	_, err := service.OwnerSummary(context.Background(), "org", "user")
	if !errors.Is(err, ErrOwnerLinkNotFound) {
		t.Fatalf("expected ErrOwnerLinkNotFound, got %v", err)
	}
}

func TestOwnerSummaryAggregatesOnlyLinkedPropertyReceivables(t *testing.T) {
	repo := fakeRepository{
		owners:      []OwnerProfile{{ID: "owner-1", LegalName: "Owner"}},
		ownerBases:  []OwnerPropertyBase{{OwnerID: "owner-1", PropertyID: "property-1", Name: "Park", OwnershipBPS: 5000}},
		receivables: []OwnerReceivableRow{{PropertyID: "property-1", Currency: "LKR", OutstandingMinor: 5000}},
	}
	service := NewService(repo, &fakeMaintenanceCreator{})
	summary, err := service.OwnerSummary(context.Background(), "org", "user")
	if err != nil {
		t.Fatal(err)
	}
	if got := summary.Properties[0].OutstandingByCurrency["LKR"]; got != 5000 {
		t.Fatalf("expected 5000 outstanding, got %d", got)
	}
}

func TestTenantSummaryRequiresExplicitTenantLink(t *testing.T) {
	service := NewService(fakeRepository{}, &fakeMaintenanceCreator{})
	_, err := service.TenantSummary(context.Background(), "org", "user")
	if !errors.Is(err, ErrTenantLinkNotFound) {
		t.Fatalf("expected ErrTenantLinkNotFound, got %v", err)
	}
}

func TestTenantMaintenanceDerivesResourceIDsFromLinkedTenancy(t *testing.T) {
	creator := &fakeMaintenanceCreator{}
	service := NewService(fakeRepository{context: TenantMaintenanceContext{TenantID: "tenant-1", PropertyID: "property-1", UnitID: "unit-1"}}, creator)
	item, err := service.CreateTenantMaintenance(context.Background(), "org", "user", CreateTenantMaintenanceInput{TenancyID: "tenancy-1", Title: "AC not cooling", Description: "Air is warm from the vents", Category: "hvac", Priority: "high"})
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "request-1" || creator.input.PropertyID != "property-1" || creator.input.UnitID != "unit-1" || creator.input.TenantID != "tenant-1" {
		t.Fatalf("portal did not derive scoped maintenance context: %+v", creator.input)
	}
}

func TestTenantMaintenanceRejectsUnlinkedTenancy(t *testing.T) {
	service := NewService(fakeRepository{contextError: ErrTenancyNotAccessible}, &fakeMaintenanceCreator{})
	_, err := service.CreateTenantMaintenance(context.Background(), "org", "user", CreateTenantMaintenanceInput{TenancyID: "other-tenancy", Title: "Issue", Description: "Description is present", Category: "other", Priority: "normal"})
	if !errors.Is(err, ErrTenancyNotAccessible) {
		t.Fatalf("expected ErrTenancyNotAccessible, got %v", err)
	}
}
