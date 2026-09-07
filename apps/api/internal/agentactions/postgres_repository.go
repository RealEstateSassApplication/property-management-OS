package agentactions

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]ActionRequest, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, organization_id, proposed_by_user_id, COALESCE(reviewed_by_user_id::text,''), action_type, resource_type, resource_id, risk_level, title, reasoning, payload, status, COALESCE(decision_reason,''), result, COALESCE(last_error,''), reviewed_at, executed_at, created_at, updated_at FROM agent_action_requests WHERE organization_id=$1 ORDER BY created_at DESC LIMIT 250`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ActionRequest, 0)
	for rows.Next() {
		item, err := scanAction(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, actionID string) (ActionRequest, error) {
	item, err := scanAction(r.pool.QueryRow(ctx, `SELECT id, organization_id, proposed_by_user_id, COALESCE(reviewed_by_user_id::text,''), action_type, resource_type, resource_id, risk_level, title, reasoning, payload, status, COALESCE(decision_reason,''), result, COALESCE(last_error,''), reviewed_at, executed_at, created_at, updated_at FROM agent_action_requests WHERE organization_id=$1 AND id=$2`, organizationID, actionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ActionRequest{}, ErrActionNotFound
	}
	return item, err
}

func (r *PostgresRepository) GetQuoteApprovalContext(ctx context.Context, organizationID, quoteID string) (QuoteApprovalContext, error) {
	var item QuoteApprovalContext
	err := r.pool.QueryRow(ctx, `SELECT q.id, q.work_order_id, wo.summary, q.vendor_id, v.name, q.amount_minor, q.currency, q.scope_summary, q.status FROM maintenance_quotes q JOIN work_orders wo ON wo.id=q.work_order_id AND wo.organization_id=q.organization_id JOIN vendors v ON v.id=q.vendor_id AND v.organization_id=q.organization_id WHERE q.organization_id=$1 AND q.id=$2`, organizationID, quoteID).Scan(&item.QuoteID, &item.WorkOrderID, &item.WorkOrderSummary, &item.VendorID, &item.VendorName, &item.AmountMinor, &item.Currency, &item.ScopeSummary, &item.QuoteStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return QuoteApprovalContext{}, ErrQuoteNotFound
	}
	return item, err
}

func (r *PostgresRepository) ProposeQuoteApproval(ctx context.Context, organizationID, proposerID, reasoning string, quote QuoteApprovalContext) (ActionRequest, error) {
	payload, err := json.Marshal(quote)
	if err != nil {
		return ActionRequest{}, err
	}
	title := "Approve maintenance quote from " + quote.VendorName
	item, err := scanAction(r.pool.QueryRow(ctx, `INSERT INTO agent_action_requests (organization_id, proposed_by_user_id, action_type, resource_type, resource_id, risk_level, title, reasoning, payload) VALUES ($1,$2,$3,'maintenance_quote',$4,'high',$5,$6,$7::jsonb) RETURNING id, organization_id, proposed_by_user_id, COALESCE(reviewed_by_user_id::text,''), action_type, resource_type, resource_id, risk_level, title, reasoning, payload, status, COALESCE(decision_reason,''), result, COALESCE(last_error,''), reviewed_at, executed_at, created_at, updated_at`, organizationID, proposerID, MaintenanceQuoteApproval, quote.QuoteID, title, reasoning, string(payload)))
	if err == nil {
		return item, nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return ActionRequest{}, err
	}
	return scanAction(r.pool.QueryRow(ctx, `SELECT id, organization_id, proposed_by_user_id, COALESCE(reviewed_by_user_id::text,''), action_type, resource_type, resource_id, risk_level, title, reasoning, payload, status, COALESCE(decision_reason,''), result, COALESCE(last_error,''), reviewed_at, executed_at, created_at, updated_at FROM agent_action_requests WHERE organization_id=$1 AND action_type=$2 AND resource_id=$3 AND status IN ('proposed','approved','executed') ORDER BY created_at DESC LIMIT 1`, organizationID, MaintenanceQuoteApproval, quote.QuoteID))
}

func (r *PostgresRepository) RecordDecision(ctx context.Context, organizationID, actionID, reviewerID, decision, reason string) (ActionRequest, error) {
	status := "rejected"
	if decision == "approve" {
		status = "approved"
	}
	item, err := scanAction(r.pool.QueryRow(ctx, `UPDATE agent_action_requests SET status=$4, reviewed_by_user_id=$3, reviewed_at=now(), decision_reason=NULLIF($5,''), last_error=NULL, updated_at=now() WHERE organization_id=$1 AND id=$2 AND status='proposed' RETURNING id, organization_id, proposed_by_user_id, COALESCE(reviewed_by_user_id::text,''), action_type, resource_type, resource_id, risk_level, title, reasoning, payload, status, COALESCE(decision_reason,''), result, COALESCE(last_error,''), reviewed_at, executed_at, created_at, updated_at`, organizationID, actionID, reviewerID, status, reason))
	if errors.Is(err, pgx.ErrNoRows) {
		if _, getErr := r.Get(ctx, organizationID, actionID); errors.Is(getErr, ErrActionNotFound) {
			return ActionRequest{}, ErrActionNotFound
		}
		return ActionRequest{}, ErrActionNotPending
	}
	return item, err
}

func (r *PostgresRepository) MarkExecuted(ctx context.Context, organizationID, actionID string, result any) (ActionRequest, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return ActionRequest{}, err
	}
	return scanAction(r.pool.QueryRow(ctx, `UPDATE agent_action_requests SET status='executed', result=$3::jsonb, executed_at=now(), last_error=NULL, updated_at=now() WHERE organization_id=$1 AND id=$2 AND status='approved' RETURNING id, organization_id, proposed_by_user_id, COALESCE(reviewed_by_user_id::text,''), action_type, resource_type, resource_id, risk_level, title, reasoning, payload, status, COALESCE(decision_reason,''), result, COALESCE(last_error,''), reviewed_at, executed_at, created_at, updated_at`, organizationID, actionID, string(data)))
}

func (r *PostgresRepository) MarkFailed(ctx context.Context, organizationID, actionID, failure string) (ActionRequest, error) {
	return scanAction(r.pool.QueryRow(ctx, `UPDATE agent_action_requests SET status='failed', last_error=$3, updated_at=now() WHERE organization_id=$1 AND id=$2 AND status='approved' RETURNING id, organization_id, proposed_by_user_id, COALESCE(reviewed_by_user_id::text,''), action_type, resource_type, resource_id, risk_level, title, reasoning, payload, status, COALESCE(decision_reason,''), result, COALESCE(last_error,''), reviewed_at, executed_at, created_at, updated_at`, organizationID, actionID, failure))
}

type rowScanner interface{ Scan(dest ...any) error }

func scanAction(row rowScanner) (ActionRequest, error) {
	var item ActionRequest
	err := row.Scan(&item.ID, &item.OrganizationID, &item.ProposedByUserID, &item.ReviewedByUserID, &item.ActionType, &item.ResourceType, &item.ResourceID, &item.RiskLevel, &item.Title, &item.Reasoning, &item.Payload, &item.Status, &item.DecisionReason, &item.Result, &item.LastError, &item.ReviewedAt, &item.ExecutedAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
