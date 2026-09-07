package agentactions

import (
	"encoding/json"
	"errors"
	"time"
)

const MaintenanceQuoteApproval = "maintenance.quote.approve"

type ActionRequest struct {
	ID               string          `json:"id"`
	OrganizationID   string          `json:"organizationId"`
	ProposedByUserID string          `json:"proposedByUserId"`
	ReviewedByUserID string          `json:"reviewedByUserId,omitempty"`
	ActionType       string          `json:"actionType"`
	ResourceType     string          `json:"resourceType"`
	ResourceID       string          `json:"resourceId"`
	RiskLevel        string          `json:"riskLevel"`
	Title            string          `json:"title"`
	Reasoning        string          `json:"reasoning"`
	Payload          json.RawMessage `json:"payload"`
	Status           string          `json:"status"`
	DecisionReason   string          `json:"decisionReason,omitempty"`
	Result           json.RawMessage `json:"result"`
	LastError        string          `json:"lastError,omitempty"`
	ReviewedAt       *time.Time      `json:"reviewedAt,omitempty"`
	ExecutedAt       *time.Time      `json:"executedAt,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

type QuoteApprovalContext struct {
	QuoteID          string `json:"quoteId"`
	WorkOrderID      string `json:"workOrderId"`
	WorkOrderSummary string `json:"workOrderSummary"`
	VendorID         string `json:"vendorId"`
	VendorName       string `json:"vendorName"`
	AmountMinor      int64  `json:"amountMinor"`
	Currency         string `json:"currency"`
	ScopeSummary     string `json:"scopeSummary"`
	QuoteStatus      string `json:"quoteStatus"`
}

type ProposeQuoteApprovalInput struct {
	QuoteID   string `json:"quoteId"`
	Reasoning string `json:"reasoning"`
}

type DecisionInput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

var (
	ErrQuoteNotFound       = errors.New("maintenance quote not found")
	ErrQuoteNotProposable  = errors.New("only submitted maintenance quotes can be proposed for approval")
	ErrActionNotFound      = errors.New("agent action request not found")
	ErrActionNotPending    = errors.New("agent action request is no longer pending review")
	ErrUnsupportedAction   = errors.New("unsupported agent action")
	ErrExecutionFailed     = errors.New("approved agent action execution failed")
)
