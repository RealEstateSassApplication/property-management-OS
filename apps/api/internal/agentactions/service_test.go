package agentactions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakeRepository struct {
	quote    QuoteApprovalContext
	proposed ActionRequest
	decision ActionRequest
	marked   ActionRequest
}

func (f *fakeRepository) List(context.Context, string) ([]ActionRequest, error) { return nil, nil }
func (f *fakeRepository) Get(context.Context, string, string) (ActionRequest, error) {
	return f.proposed, nil
}
func (f *fakeRepository) GetQuoteApprovalContext(context.Context, string, string) (QuoteApprovalContext, error) {
	return f.quote, nil
}
func (f *fakeRepository) ProposeQuoteApproval(_ context.Context, _, proposerID, reasoning string, quote QuoteApprovalContext) (ActionRequest, error) {
	payload, _ := json.Marshal(quote)
	f.proposed = ActionRequest{ID: "action-1", ProposedByUserID: proposerID, ActionType: MaintenanceQuoteApproval, ResourceID: quote.QuoteID, Reasoning: reasoning, Payload: payload, Status: "proposed"}
	return f.proposed, nil
}
func (f *fakeRepository) RecordDecision(_ context.Context, _, _, reviewerID, decision, reason string) (ActionRequest, error) {
	f.decision = f.proposed
	f.decision.ReviewedByUserID = reviewerID
	f.decision.DecisionReason = reason
	if decision == "approve" {
		f.decision.Status = "approved"
	} else {
		f.decision.Status = "rejected"
	}
	return f.decision, nil
}
func (f *fakeRepository) MarkExecuted(context.Context, string, string, any) (ActionRequest, error) {
	f.marked = f.decision
	f.marked.Status = "executed"
	return f.marked, nil
}
func (f *fakeRepository) MarkFailed(_ context.Context, _, _, failure string) (ActionRequest, error) {
	f.marked = f.decision
	f.marked.Status = "failed"
	f.marked.LastError = failure
	return f.marked, nil
}

type fakeExecutor struct{ err error }

func (f fakeExecutor) ApproveMaintenanceQuote(context.Context, string, string, string) (any, error) {
	if f.err != nil {
		return nil, f.err
	}
	return map[string]any{"status": "approved"}, nil
}

func TestProposeQuoteApprovalSnapshotsSubmittedQuote(t *testing.T) {
	repo := &fakeRepository{quote: QuoteApprovalContext{QuoteID: "quote-1", VendorName: "Vendor", QuoteStatus: "submitted", AmountMinor: 1000, Currency: "LKR"}}
	service := NewService(repo, fakeExecutor{})
	item, err := service.ProposeMaintenanceQuoteApproval(context.Background(), "org", "agent-user", ProposeQuoteApprovalInput{QuoteID: "quote-1", Reasoning: "Lowest valid quote for the urgent repair."})
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != "proposed" || item.ProposedByUserID != "agent-user" || item.ResourceID != "quote-1" {
		t.Fatalf("unexpected proposal: %+v", item)
	}
}

func TestProposeRejectsNonSubmittedQuote(t *testing.T) {
	repo := &fakeRepository{quote: QuoteApprovalContext{QuoteID: "quote-1", QuoteStatus: "approved"}}
	service := NewService(repo, fakeExecutor{})
	_, err := service.ProposeMaintenanceQuoteApproval(context.Background(), "org", "agent-user", ProposeQuoteApprovalInput{QuoteID: "quote-1", Reasoning: "This reasoning is sufficiently descriptive."})
	if !errors.Is(err, ErrQuoteNotProposable) {
		t.Fatalf("expected ErrQuoteNotProposable, got %v", err)
	}
}

func TestHumanApprovalExecutesDomainAction(t *testing.T) {
	payload, _ := json.Marshal(QuoteApprovalContext{QuoteID: "quote-1"})
	repo := &fakeRepository{proposed: ActionRequest{ID: "action-1", ActionType: MaintenanceQuoteApproval, Payload: payload, Status: "proposed"}}
	service := NewService(repo, fakeExecutor{})
	item, err := service.Decide(context.Background(), "org", "action-1", "reviewer", DecisionInput{Decision: "approve", Reason: "Budget and scope are acceptable."})
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != "executed" {
		t.Fatalf("expected executed, got %s", item.Status)
	}
}

func TestExecutionFailureIsRecorded(t *testing.T) {
	payload, _ := json.Marshal(QuoteApprovalContext{QuoteID: "quote-1"})
	repo := &fakeRepository{proposed: ActionRequest{ID: "action-1", ActionType: MaintenanceQuoteApproval, Payload: payload, Status: "proposed"}}
	service := NewService(repo, fakeExecutor{err: errors.New("quote changed")})
	item, err := service.Decide(context.Background(), "org", "action-1", "reviewer", DecisionInput{Decision: "approve", Reason: "Approve the submitted quote."})
	if !errors.Is(err, ErrExecutionFailed) || item.Status != "failed" {
		t.Fatalf("expected recorded execution failure, item=%+v err=%v", item, err)
	}
}
