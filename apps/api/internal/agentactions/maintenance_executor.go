package agentactions

import (
	"context"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/maintenance"
)

type MaintenanceExecutor struct{ service *maintenance.Service }

func NewMaintenanceExecutor(service *maintenance.Service) *MaintenanceExecutor {
	return &MaintenanceExecutor{service: service}
}

func (e *MaintenanceExecutor) ApproveMaintenanceQuote(ctx context.Context, organizationID, quoteID, reviewerID string) (any, error) {
	quote, err := e.service.DecideQuote(ctx, organizationID, quoteID, reviewerID, maintenance.DecideQuoteInput{Decision: "approve"})
	if err == nil {
		return quote, nil
	}

	// If a process crashed after the domain action succeeded but before the
	// proposal was marked executed, treat an already-approved quote as a
	// successful replay rather than issuing the financial decision twice.
	quotes, listErr := e.service.ListQuotes(ctx, organizationID)
	if listErr != nil {
		return nil, err
	}
	for _, existing := range quotes {
		if existing.ID == quoteID && existing.Status == "approved" {
			return existing, nil
		}
	}
	return nil, err
}
