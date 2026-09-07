package maintenance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListVendors(ctx context.Context, organizationID string) ([]Vendor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, organization_id, name, trade, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
		FROM vendors WHERE organization_id = $1
		ORDER BY status, name
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Vendor, 0)
	for rows.Next() {
		var item Vendor
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Name, &item.Trade, &item.Email, &item.Phone, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateVendor(ctx context.Context, organizationID string, input CreateVendorInput) (Vendor, error) {
	var id string
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO vendors (organization_id, name, trade, email, phone, status)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6)
		RETURNING id
	`, organizationID, input.Name, input.Trade, input.Email, input.Phone, input.Status).Scan(&id); err != nil {
		return Vendor{}, err
	}
	var item Vendor
	err := r.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, trade, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
		FROM vendors WHERE organization_id = $1 AND id = $2
	`, organizationID, id).Scan(&item.ID, &item.OrganizationID, &item.Name, &item.Trade, &item.Email, &item.Phone, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

const requestSelect = `
	SELECT mr.id, mr.organization_id, mr.property_id, p.name,
		COALESCE(mr.unit_id::text, ''), COALESCE(u.label, ''),
		COALESCE(mr.tenant_id::text, ''), COALESCE(t.legal_name, ''),
		mr.title, mr.description, mr.category, mr.priority, mr.status,
		mr.resolved_at, mr.created_at, mr.updated_at
	FROM maintenance_requests mr
	JOIN properties p ON p.id = mr.property_id AND p.organization_id = mr.organization_id
	LEFT JOIN units u ON u.id = mr.unit_id AND u.organization_id = mr.organization_id
	LEFT JOIN tenants t ON t.id = mr.tenant_id AND t.organization_id = mr.organization_id
`

func (r *PostgresRepository) ListRequests(ctx context.Context, organizationID string) ([]Request, error) {
	rows, err := r.pool.Query(ctx, requestSelect+`
		WHERE mr.organization_id = $1
		ORDER BY CASE mr.priority WHEN 'emergency' THEN 1 WHEN 'high' THEN 2 WHEN 'normal' THEN 3 ELSE 4 END, mr.created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Request, 0)
	for rows.Next() {
		var item Request
		if err := scanRequest(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateRequest(ctx context.Context, organizationID, actorUserID string, input CreateRequestInput) (Request, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Request{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var propertyExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM properties WHERE organization_id = $1 AND id = $2)`, organizationID, input.PropertyID).Scan(&propertyExists); err != nil {
		return Request{}, err
	}
	if !propertyExists {
		return Request{}, ErrPropertyNotFound
	}
	if input.UnitID != "" {
		var unitExists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM units WHERE organization_id = $1 AND id = $2 AND property_id = $3)`, organizationID, input.UnitID, input.PropertyID).Scan(&unitExists); err != nil {
			return Request{}, err
		}
		if !unitExists {
			return Request{}, ErrUnitNotFound
		}
	}
	if input.TenantID != "" {
		if input.UnitID == "" {
			return Request{}, ErrTenantNotOccupant
		}
		var tenantExists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenants WHERE organization_id = $1 AND id = $2)`, organizationID, input.TenantID).Scan(&tenantExists); err != nil {
			return Request{}, err
		}
		if !tenantExists {
			return Request{}, ErrTenantNotFound
		}
		var activeOccupant bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM tenancies tenancy
				JOIN tenancy_tenants link ON link.tenancy_id = tenancy.id AND link.organization_id = tenancy.organization_id
				WHERE tenancy.organization_id = $1 AND tenancy.unit_id = $2 AND tenancy.status = 'active' AND link.tenant_id = $3
			)
		`, organizationID, input.UnitID, input.TenantID).Scan(&activeOccupant); err != nil {
			return Request{}, err
		}
		if !activeOccupant {
			return Request{}, ErrTenantNotOccupant
		}
	}

	var id string
	if err := tx.QueryRow(ctx, `
		INSERT INTO maintenance_requests (
			organization_id, property_id, unit_id, tenant_id, reported_by_user_id,
			title, description, category, priority
		) VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, NULLIF($5, '')::uuid, $6, $7, $8, $9)
		RETURNING id
	`, organizationID, input.PropertyID, input.UnitID, input.TenantID, actorUserID, input.Title, input.Description, input.Category, input.Priority).Scan(&id); err != nil {
		return Request{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Request{}, err
	}
	return r.getRequest(ctx, organizationID, id)
}

func (r *PostgresRepository) UpdateRequestStatus(ctx context.Context, organizationID, requestID, target string) (Request, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Request{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var current string
	if err := tx.QueryRow(ctx, `SELECT status FROM maintenance_requests WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, requestID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Request{}, ErrRequestNotFound
		}
		return Request{}, err
	}
	if !validRequestTransition(current, target) {
		return Request{}, ErrInvalidRequestTransition
	}
	if target == "resolved" {
		var incomplete, completed int
		if err := tx.QueryRow(ctx, `
			SELECT count(*) FILTER (WHERE status IN ('planned','assigned','in_progress')),
			       count(*) FILTER (WHERE status = 'completed')
			FROM work_orders WHERE organization_id = $1 AND maintenance_request_id = $2
		`, organizationID, requestID).Scan(&incomplete, &completed); err != nil {
			return Request{}, err
		}
		if incomplete > 0 || completed == 0 {
			return Request{}, ErrInvalidRequestTransition
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE maintenance_requests
		SET status = $3, resolved_at = CASE WHEN $3 = 'resolved' THEN now() ELSE resolved_at END, updated_at = now()
		WHERE organization_id = $1 AND id = $2
	`, organizationID, requestID, target); err != nil {
		return Request{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Request{}, err
	}
	return r.getRequest(ctx, organizationID, requestID)
}

func (r *PostgresRepository) getRequest(ctx context.Context, organizationID, requestID string) (Request, error) {
	var item Request
	err := scanRequest(r.pool.QueryRow(ctx, requestSelect+` WHERE mr.organization_id = $1 AND mr.id = $2`, organizationID, requestID), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return Request{}, ErrRequestNotFound
	}
	return item, err
}

const workOrderSelect = `
	SELECT wo.id, wo.organization_id, wo.maintenance_request_id, mr.title, p.name, COALESCE(u.label, ''),
		COALESCE(wo.vendor_id::text, ''), COALESCE(v.name, ''), COALESCE(wo.assigned_user_id::text, ''),
		wo.summary, wo.status, wo.scheduled_for, wo.started_at, wo.completed_at, wo.created_at, wo.updated_at
	FROM work_orders wo
	JOIN maintenance_requests mr ON mr.id = wo.maintenance_request_id AND mr.organization_id = wo.organization_id
	JOIN properties p ON p.id = mr.property_id AND p.organization_id = wo.organization_id
	LEFT JOIN units u ON u.id = mr.unit_id AND u.organization_id = wo.organization_id
	LEFT JOIN vendors v ON v.id = wo.vendor_id AND v.organization_id = wo.organization_id
`

func (r *PostgresRepository) ListWorkOrders(ctx context.Context, organizationID string) ([]WorkOrder, error) {
	rows, err := r.pool.Query(ctx, workOrderSelect+` WHERE wo.organization_id = $1 ORDER BY wo.created_at DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]WorkOrder, 0)
	for rows.Next() {
		var item WorkOrder
		if err := scanWorkOrder(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateWorkOrder(ctx context.Context, organizationID string, input CreateWorkOrderInput, scheduledFor *time.Time) (WorkOrder, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return WorkOrder{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var requestStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM maintenance_requests WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, input.MaintenanceRequestID).Scan(&requestStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkOrder{}, ErrRequestNotFound
		}
		return WorkOrder{}, err
	}
	if requestStatus == "resolved" || requestStatus == "cancelled" {
		return WorkOrder{}, ErrRequestClosed
	}
	if input.VendorID != "" {
		if err := validateActiveVendor(ctx, tx, organizationID, input.VendorID); err != nil {
			return WorkOrder{}, err
		}
	}
	if input.AssignedUserID != "" {
		var member bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM organization_memberships WHERE organization_id = $1 AND user_id = $2)`, organizationID, input.AssignedUserID).Scan(&member); err != nil {
			return WorkOrder{}, err
		}
		if !member {
			return WorkOrder{}, fmt.Errorf("assigned user is not an organization member")
		}
	}
	status := "planned"
	if input.VendorID != "" || input.AssignedUserID != "" {
		status = "assigned"
	}
	var id string
	if err := tx.QueryRow(ctx, `
		INSERT INTO work_orders (organization_id, maintenance_request_id, vendor_id, assigned_user_id, summary, status, scheduled_for)
		VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, $5, $6, $7)
		RETURNING id
	`, organizationID, input.MaintenanceRequestID, input.VendorID, input.AssignedUserID, input.Summary, status, scheduledFor).Scan(&id); err != nil {
		return WorkOrder{}, err
	}
	if requestStatus == "open" {
		if _, err := tx.Exec(ctx, `UPDATE maintenance_requests SET status = 'triaged', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, input.MaintenanceRequestID); err != nil {
			return WorkOrder{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkOrder{}, err
	}
	return r.getWorkOrder(ctx, organizationID, id)
}

func (r *PostgresRepository) UpdateWorkOrderStatus(ctx context.Context, organizationID, workOrderID, target string) (WorkOrder, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return WorkOrder{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var current, requestID string
	if err := tx.QueryRow(ctx, `SELECT status, maintenance_request_id FROM work_orders WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, workOrderID).Scan(&current, &requestID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkOrder{}, ErrWorkOrderNotFound
		}
		return WorkOrder{}, err
	}
	if !validWorkOrderTransition(current, target) {
		return WorkOrder{}, ErrInvalidWorkOrderTransition
	}
	if target == "completed" {
		var evidenceCount int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM maintenance_completion_evidence WHERE organization_id = $1 AND work_order_id = $2`, organizationID, workOrderID).Scan(&evidenceCount); err != nil {
			return WorkOrder{}, err
		}
		if evidenceCount == 0 {
			return WorkOrder{}, ErrCompletionEvidenceRequired
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE work_orders SET status = $3,
			started_at = CASE WHEN $3 = 'in_progress' AND started_at IS NULL THEN now() ELSE started_at END,
			completed_at = CASE WHEN $3 = 'completed' THEN now() ELSE completed_at END,
			updated_at = now()
		WHERE organization_id = $1 AND id = $2
	`, organizationID, workOrderID, target); err != nil {
		return WorkOrder{}, err
	}
	if target == "in_progress" {
		if _, err := tx.Exec(ctx, `UPDATE maintenance_requests SET status = 'in_progress', updated_at = now() WHERE organization_id = $1 AND id = $2 AND status IN ('open','triaged')`, organizationID, requestID); err != nil {
			return WorkOrder{}, err
		}
	}
	if target == "completed" {
		var incomplete, completed int
		if err := tx.QueryRow(ctx, `
			SELECT count(*) FILTER (WHERE status IN ('planned','assigned','in_progress')),
			       count(*) FILTER (WHERE status = 'completed')
			FROM work_orders WHERE organization_id = $1 AND maintenance_request_id = $2
		`, organizationID, requestID).Scan(&incomplete, &completed); err != nil {
			return WorkOrder{}, err
		}
		if incomplete == 0 && completed > 0 {
			if _, err := tx.Exec(ctx, `UPDATE maintenance_requests SET status = 'resolved', resolved_at = now(), updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, requestID); err != nil {
				return WorkOrder{}, err
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkOrder{}, err
	}
	return r.getWorkOrder(ctx, organizationID, workOrderID)
}

func (r *PostgresRepository) getWorkOrder(ctx context.Context, organizationID, id string) (WorkOrder, error) {
	var item WorkOrder
	err := scanWorkOrder(r.pool.QueryRow(ctx, workOrderSelect+` WHERE wo.organization_id = $1 AND wo.id = $2`, organizationID, id), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkOrder{}, ErrWorkOrderNotFound
	}
	return item, err
}

const quoteSelect = `
	SELECT q.id, q.organization_id, q.work_order_id, wo.summary, q.vendor_id, v.name,
		q.amount_minor, q.currency, q.scope_summary, q.status, q.submitted_at, q.reviewed_at,
		COALESCE(q.reviewed_by_user_id::text, ''), q.created_at, q.updated_at
	FROM maintenance_quotes q
	JOIN work_orders wo ON wo.id = q.work_order_id AND wo.organization_id = q.organization_id
	JOIN vendors v ON v.id = q.vendor_id AND v.organization_id = q.organization_id
`

func (r *PostgresRepository) ListQuotes(ctx context.Context, organizationID string) ([]Quote, error) {
	rows, err := r.pool.Query(ctx, quoteSelect+` WHERE q.organization_id = $1 ORDER BY q.submitted_at DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Quote, 0)
	for rows.Next() {
		var item Quote
		if err := scanQuote(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateQuote(ctx context.Context, organizationID string, input CreateQuoteInput) (Quote, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Quote{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var workOrderStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM work_orders WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, input.WorkOrderID).Scan(&workOrderStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Quote{}, ErrWorkOrderNotFound
		}
		return Quote{}, err
	}
	if workOrderStatus == "completed" || workOrderStatus == "cancelled" {
		return Quote{}, ErrWorkOrderClosed
	}
	if err := validateActiveVendor(ctx, tx, organizationID, input.VendorID); err != nil {
		return Quote{}, err
	}
	var id string
	if err := tx.QueryRow(ctx, `
		INSERT INTO maintenance_quotes (organization_id, work_order_id, vendor_id, amount_minor, currency, scope_summary)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`, organizationID, input.WorkOrderID, input.VendorID, input.AmountMinor, input.Currency, input.ScopeSummary).Scan(&id); err != nil {
		return Quote{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Quote{}, err
	}
	return r.getQuote(ctx, organizationID, id)
}

func (r *PostgresRepository) DecideQuote(ctx context.Context, organizationID, quoteID, actorUserID, decision string) (Quote, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Quote{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status, workOrderID, vendorID string
	if err := tx.QueryRow(ctx, `
		SELECT status, work_order_id, vendor_id FROM maintenance_quotes
		WHERE organization_id = $1 AND id = $2 FOR UPDATE
	`, organizationID, quoteID).Scan(&status, &workOrderID, &vendorID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Quote{}, ErrQuoteNotFound
		}
		return Quote{}, err
	}
	if status != "submitted" {
		return Quote{}, ErrQuoteNotSubmitted
	}
	var workOrderStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM work_orders WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, workOrderID).Scan(&workOrderStatus); err != nil {
		return Quote{}, err
	}
	if workOrderStatus == "completed" || workOrderStatus == "cancelled" {
		return Quote{}, ErrWorkOrderClosed
	}
	newStatus := "rejected"
	if decision == "approve" {
		var approvedExists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM maintenance_quotes WHERE organization_id = $1 AND work_order_id = $2 AND status = 'approved' AND id <> $3)`, organizationID, workOrderID, quoteID).Scan(&approvedExists); err != nil {
			return Quote{}, err
		}
		if approvedExists {
			return Quote{}, ErrApprovedQuoteExists
		}
		newStatus = "approved"
		if _, err := tx.Exec(ctx, `
			UPDATE work_orders SET vendor_id = $3,
				status = CASE WHEN status = 'planned' THEN 'assigned' ELSE status END,
				updated_at = now()
			WHERE organization_id = $1 AND id = $2
		`, organizationID, workOrderID, vendorID); err != nil {
			return Quote{}, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE maintenance_quotes SET status = 'rejected', reviewed_at = now(), reviewed_by_user_id = NULLIF($4, '')::uuid, updated_at = now()
			WHERE organization_id = $1 AND work_order_id = $2 AND id <> $3 AND status = 'submitted'
		`, organizationID, workOrderID, quoteID, actorUserID); err != nil {
			return Quote{}, err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE maintenance_quotes SET status = $3, reviewed_at = now(), reviewed_by_user_id = NULLIF($4, '')::uuid, updated_at = now()
		WHERE organization_id = $1 AND id = $2
	`, organizationID, quoteID, newStatus, actorUserID); err != nil {
		return Quote{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events (organization_id, actor_user_id, action, resource_type, resource_id, metadata)
		VALUES ($1, NULLIF($2, '')::uuid, $3, 'maintenance_quote', $4, jsonb_build_object('decision', $5, 'workOrderId', $6))
	`, organizationID, actorUserID, "maintenance.quote."+newStatus, quoteID, decision, workOrderID); err != nil {
		return Quote{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Quote{}, err
	}
	return r.getQuote(ctx, organizationID, quoteID)
}

func (r *PostgresRepository) getQuote(ctx context.Context, organizationID, id string) (Quote, error) {
	var item Quote
	err := scanQuote(r.pool.QueryRow(ctx, quoteSelect+` WHERE q.organization_id = $1 AND q.id = $2`, organizationID, id), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return Quote{}, ErrQuoteNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListEvidence(ctx context.Context, organizationID string) ([]Evidence, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, organization_id, work_order_id, evidence_type, COALESCE(note, ''), COALESCE(storage_key, ''), COALESCE(submitted_by_user_id::text, ''), created_at
		FROM maintenance_completion_evidence WHERE organization_id = $1 ORDER BY created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Evidence, 0)
	for rows.Next() {
		var item Evidence
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.WorkOrderID, &item.EvidenceType, &item.Note, &item.StorageKey, &item.SubmittedByUserID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateEvidence(ctx context.Context, organizationID, actorUserID string, input CreateEvidenceInput) (Evidence, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Evidence{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var workOrderStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM work_orders WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, input.WorkOrderID).Scan(&workOrderStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Evidence{}, ErrWorkOrderNotFound
		}
		return Evidence{}, err
	}
	if workOrderStatus == "cancelled" {
		return Evidence{}, ErrWorkOrderClosed
	}
	var item Evidence
	if err := tx.QueryRow(ctx, `
		INSERT INTO maintenance_completion_evidence (organization_id, work_order_id, evidence_type, note, storage_key, submitted_by_user_id)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, '')::uuid)
		RETURNING id, organization_id, work_order_id, evidence_type, COALESCE(note, ''), COALESCE(storage_key, ''), COALESCE(submitted_by_user_id::text, ''), created_at
	`, organizationID, input.WorkOrderID, input.EvidenceType, input.Note, input.StorageKey, actorUserID).Scan(
		&item.ID, &item.OrganizationID, &item.WorkOrderID, &item.EvidenceType, &item.Note, &item.StorageKey, &item.SubmittedByUserID, &item.CreatedAt,
	); err != nil {
		return Evidence{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events (organization_id, actor_user_id, action, resource_type, resource_id, metadata)
		VALUES ($1, NULLIF($2, '')::uuid, 'maintenance.evidence.created', 'work_order', $3, jsonb_build_object('evidenceId', $4, 'type', $5))
	`, organizationID, actorUserID, input.WorkOrderID, item.ID, input.EvidenceType); err != nil {
		return Evidence{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Evidence{}, err
	}
	return item, nil
}

func validateActiveVendor(ctx context.Context, tx pgx.Tx, organizationID, vendorID string) error {
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM vendors WHERE organization_id = $1 AND id = $2`, organizationID, vendorID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrVendorNotFound
		}
		return err
	}
	if status != "active" {
		return ErrVendorInactive
	}
	return nil
}

func validRequestTransition(current, target string) bool {
	if current == target {
		return true
	}
	switch current {
	case "open":
		return target == "triaged" || target == "cancelled"
	case "triaged":
		return target == "in_progress" || target == "cancelled"
	case "in_progress":
		return target == "resolved" || target == "cancelled"
	default:
		return false
	}
}

func validWorkOrderTransition(current, target string) bool {
	if current == target {
		return true
	}
	switch current {
	case "planned":
		return target == "assigned" || target == "in_progress" || target == "cancelled"
	case "assigned":
		return target == "in_progress" || target == "cancelled"
	case "in_progress":
		return target == "completed" || target == "cancelled"
	default:
		return false
	}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRequest(row scanner, item *Request) error {
	return row.Scan(
		&item.ID, &item.OrganizationID, &item.PropertyID, &item.PropertyName,
		&item.UnitID, &item.UnitLabel, &item.TenantID, &item.TenantName,
		&item.Title, &item.Description, &item.Category, &item.Priority, &item.Status,
		&item.ResolvedAt, &item.CreatedAt, &item.UpdatedAt,
	)
}

func scanWorkOrder(row scanner, item *WorkOrder) error {
	return row.Scan(
		&item.ID, &item.OrganizationID, &item.MaintenanceRequestID, &item.RequestTitle,
		&item.PropertyName, &item.UnitLabel, &item.VendorID, &item.VendorName,
		&item.AssignedUserID, &item.Summary, &item.Status, &item.ScheduledFor,
		&item.StartedAt, &item.CompletedAt, &item.CreatedAt, &item.UpdatedAt,
	)
}

func scanQuote(row scanner, item *Quote) error {
	return row.Scan(
		&item.ID, &item.OrganizationID, &item.WorkOrderID, &item.WorkOrderSummary,
		&item.VendorID, &item.VendorName, &item.AmountMinor, &item.Currency,
		&item.ScopeSummary, &item.Status, &item.SubmittedAt, &item.ReviewedAt,
		&item.ReviewedByUserID, &item.CreatedAt, &item.UpdatedAt,
	)
}
