package agentactions

import (
	"context"
	"encoding/json"
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
	if decision != "approve" && decision != "reject" {
		return ActionRequest{}, fmt.Errorf("decision must be approve or reject")
	}
	if reason == "" {
		return ActionRequest{}, fmt.Errorf("decision reason is required")
	}
	action, err := s.repository.RecordDecision(ctx, organizationID, strings.TrimSpace(actionID), reviewerID, decision, reason)
	if err != nil || decision == "reject" {
		return action, err
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
	result, err := s.executor.ApproveMaintenanceQuote(ctx, organizationID, payload.QuoteID, reviewerID)
	if err != nil {
		failed, markErr := s.repository.MarkFailed(ctx, organizationID, action.ID, err.Error())
		if markErr != nil {
			return ActionRequest{}, markErr
		}
		return failed, fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}
	return s.repository.MarkExecuted(ctx, organizationID, action.ID, result)
}

func (s *Service) failUnsupported(ctx context.Context, organizationID string, action ActionRequest) (ActionRequest, error) {
	failed, err := s.repository.MarkFailed(ctx, organizationID, action.ID, ErrUnsupportedAction.Error())
	if err != nil {
		return ActionRequest{}, err
	}
	return failed, ErrUnsupportedAction
}
