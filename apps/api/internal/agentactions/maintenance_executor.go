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

	// If a process crashed after this human reviewer's domain action succeeded
	// but before the proposal result was persisted, an already-approved quote by
	// the same reviewer is a safe replay. Approval by another actor is treated as
	// stale state and must not be claimed as this proposal's execution.
	quotes, listErr := e.service.ListQuotes(ctx, organizationID)
	if listErr != nil {
		return nil, err
	}
	for _, existing := range quotes {
		if existing.ID == quoteID && existing.Status == "approved" && existing.ReviewedByUserID == reviewerID {
			return existing, nil
		}
	}
	return nil, err
}
