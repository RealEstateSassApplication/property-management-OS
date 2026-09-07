package reporting

import (
	"context"
	"testing"
)

type fakeReportingRepository struct {
	dashboard Dashboard
	limit     int
}

func (f *fakeReportingRepository) Dashboard(context.Context, string) (Dashboard, error) {
	return f.dashboard, nil
}
func (f *fakeReportingRepository) ListAudit(_ context.Context, _ string, limit int) ([]AuditEvent, error) {
	f.limit = limit
	return nil, nil
}

func TestDashboardReturnsDerivedSnapshot(t *testing.T) {
	repo := &fakeReportingRepository{dashboard: Dashboard{PropertyCount: 2, UnitCount: 10, OccupiedUnits: 8, VacancyRateBPS: 2000}}
	service := NewService(repo)
	item, err := service.Dashboard(context.Background(), "org")
	if err != nil {
		t.Fatal(err)
	}
	if item.PropertyCount != 2 || item.UnitCount != 10 || item.OccupiedUnits != 8 || item.VacancyRateBPS != 2000 {
		t.Fatalf("unexpected dashboard: %+v", item)
	}
}

func TestAuditLimitDefaultsAndCaps(t *testing.T) {
	repo := &fakeReportingRepository{}
	service := NewService(repo)
	if _, err := service.ListAudit(context.Background(), "org", 0); err != nil {
		t.Fatal(err)
	}
	if repo.limit != 100 {
		t.Fatalf("expected default 100, got %d", repo.limit)
	}
	if _, err := service.ListAudit(context.Background(), "org", 9999); err != nil {
		t.Fatal(err)
	}
	if repo.limit != 500 {
		t.Fatalf("expected cap 500, got %d", repo.limit)
	}
}
