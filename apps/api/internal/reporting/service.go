package reporting

import "context"

type Repository interface {
	Dashboard(ctx context.Context, organizationID string) (Dashboard, error)
	ListAudit(ctx context.Context, organizationID string, limit int) ([]AuditEvent, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Dashboard(ctx context.Context, organizationID string) (Dashboard, error) {
	return s.repository.Dashboard(ctx, organizationID)
}

func (s *Service) ListAudit(ctx context.Context, organizationID string, limit int) ([]AuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	return s.repository.ListAudit(ctx, organizationID, limit)
}
