package agentactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Repository interface {
	List(ctx context.Context, organizationID string) ([]ActionRequest, error)
	Get(ctx context.Context, organizationID, actionID string) (ActionRequest, error)
	GetQuoteApprovalContext(ctx context.Context, organizationID, quoteID string) (QuoteApprovalContext, error)
	ProposeQuoteApproval(ctx context.Context, organizationID, proposerID, reasoning string, quote QuoteApprovalContext) (ActionRequest, error)
	RecordDecision(ctx context.Context, organizationID, actionID, reviewerID, decision, reason string) (ActionRequest, error)
	MarkExecuted(ctx context.Context, organizationID, actionID string, result any) (ActionRequest, error)
	MarkFailed(ctx context.Context, organizationID, actionID, failure string) (ActionRequest, error)
}

type Executor interface {
	ApproveMaintenanceQuote(ctx context.Context, organizationID, quoteID, reviewerID string) (any, error)
}

type Service struct {
	repository Repository
	executor   Executor
}

func NewService(repository Repository, executor Executor) *Service {
	return &Service{repository: repository, executor: executor}
}

func (s *Service) List(ctx context.Context, organizationID string) ([]ActionRequest, error) {
	return s.repository.List(ctx, organizationID)
}

func (s *Service) ProposeMaintenanceQuoteApproval(ctx context.Context, organizationID, proposerID string, input ProposeQuoteApprovalInput) (ActionRequest, error) {
	input.QuoteID = strings.TrimSpace(input.QuoteID)
	input.Reasoning = strings.TrimSpace(input.Reasoning)
	if input.QuoteID == "" {
		return ActionRequest{}, fmt.Errorf("quoteId is required")
	}
	if len(input.Reasoning) < 12 {
		return ActionRequest{}, fmt.Errorf("reasoning must explain why approval is recommended")
	}
	quote, err := s.repository.GetQuoteApprovalContext(ctx, organizationID, input.QuoteID)
	if err != nil {
		return ActionRequest{}, err
	}
	if quote.QuoteStatus != "submitted" {
		return ActionRequest{}, ErrQuoteNotProposable
	}
	return s.repository.ProposeQuoteApproval(ctx, organizationID, proposerID, input.Reasoning, quote)
}

func (s *Service) Decide(ctx context.Context, organizationID, actionID, reviewerID string, input DecisionInput) (ActionRequest, error) {
	decision := strings.ToLower(strings.TrimSpace(input.Decision))
	reason := strings.TrimSpace(input.Reason)
	actionID = strings.TrimSpace(actionID)
	reviewerID = strings.TrimSpace(reviewerID)
	if decision != "approve" && decision != "reject" {
		return ActionRequest{}, fmt.Errorf("decision must be approve or reject")
	}
	if reason == "" {
		return ActionRequest{}, fmt.Errorf("decision reason is required")
	}

	action, err := s.repository.Get(ctx, organizationID, actionID)
	if err != nil {
		return ActionRequest{}, err
	}

	switch action.Status {
	case "proposed":
		action, err = s.repository.RecordDecision(ctx, organizationID, actionID, reviewerID, decision, reason)
		if errors.Is(err, ErrActionNotPending) && decision == "approve" {
			// Another request may have recorded the same approval between Get and
			// RecordDecision. Reload and continue only if the durable state is approved.
			action, err = s.repository.Get(ctx, organizationID, actionID)
			if err == nil && action.Status != "approved" {
				return ActionRequest{}, ErrActionNotPending
			}
		}
		if err != nil {
			return action, err
		}
	case "approved":
		// Recovery path for a process that stopped after the human decision was
		// committed but before domain execution/result persistence completed.
		if decision != "approve" {
			return ActionRequest{}, ErrActionNotPending
		}
	case "executed":
		if decision == "approve" {
			return action, nil
		}
		return ActionRequest{}, ErrActionNotPending
	default:
		return ActionRequest{}, ErrActionNotPending
	}

	if decision == "reject" {
		return action, nil
	}
	if action.ActionType != MaintenanceQuoteApproval {
		return s.failUnsupported(ctx, organizationID, action)
	}

	var payload QuoteApprovalContext
	if err := json.Unmarshal(action.Payload, &payload); err != nil {
		failed, markErr := s.repository.MarkFailed(ctx, organizationID, action.ID, "invalid action payload")
		if markErr != nil {
			return ActionRequest{}, markErr
		}
		return failed, fmt.Errorf("%w: invalid action payload", ErrExecutionFailed)
	}

	executionReviewerID := reviewerID
	if action.ReviewedByUserID != "" {
		executionReviewerID = action.ReviewedByUserID
	}
	result, err := s.executor.ApproveMaintenanceQuote(ctx, organizationID, payload.QuoteID, executionReviewerID)
	if err != nil {
		failed, markErr := s.repository.MarkFailed(ctx, organizationID, action.ID, err.Error())
		if markErr != nil {
			return ActionRequest{}, markErr
		}
		return failed, fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}

	executed, err := s.repository.MarkExecuted(ctx, organizationID, action.ID, result)
	if err == nil {
		return executed, nil
	}
	// Concurrent recovery may have persisted the result first. Do not turn an
	// already executed action into a false failure.
	current, getErr := s.repository.Get(ctx, organizationID, action.ID)
	if getErr == nil && current.Status == "executed" {
		return current, nil
	}
	return ActionRequest{}, err
}

func (s *Service) failUnsupported(ctx context.Context, organizationID string, action ActionRequest) (ActionRequest, error) {
	failed, err := s.repository.MarkFailed(ctx, organizationID, action.ID, ErrUnsupportedAction.Error())
	if err != nil {
		return ActionRequest{}, err
	}
	return failed, ErrUnsupportedAction
}
