package accounting

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const expenseSelect = `
	SELECT e.id,e.organization_id,e.property_id,p.name,COALESCE(e.vendor_id::text,''),COALESCE(v.name,''),
	       COALESCE(e.work_order_id::text,''),e.category,e.amount_minor,e.currency,to_char(e.incurred_on,'YYYY-MM-DD'),e.note,
	       COALESCE(e.reference_code,''),COALESCE(e.created_by_user_id::text,''),e.created_at,
	       (er.id IS NOT NULL),COALESCE(er.reason,''),er.reversed_at
	FROM property_expenses e
	JOIN properties p ON p.id=e.property_id AND p.organization_id=e.organization_id
	LEFT JOIN vendors v ON v.id=e.vendor_id AND v.organization_id=e.organization_id
	LEFT JOIN property_expense_reversals er ON er.expense_id=e.id AND er.organization_id=e.organization_id
`

func (r *PostgresRepository) ListExpenses(ctx context.Context, organizationID string) ([]PropertyExpense, error) {
	rows, err := r.pool.Query(ctx, expenseSelect+` WHERE e.organization_id=$1 ORDER BY e.incurred_on DESC,e.created_at DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PropertyExpense, 0)
	for rows.Next() {
		var item PropertyExpense
		if err := scanExpense(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateExpense(ctx context.Context, organizationID, actorUserID string, input CreateExpenseInput) (PropertyExpense, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return PropertyExpense{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var propertyExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM properties WHERE organization_id=$1 AND id=$2 AND status<>'archived')`, organizationID, input.PropertyID).Scan(&propertyExists); err != nil {
		return PropertyExpense{}, err
	}
	if !propertyExists {
		return PropertyExpense{}, ErrPropertyNotFound
	}
	if input.VendorID != "" {
		var vendorExists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vendors WHERE organization_id=$1 AND id=$2)`, organizationID, input.VendorID).Scan(&vendorExists); err != nil {
			return PropertyExpense{}, err
		}
		if !vendorExists {
			return PropertyExpense{}, ErrVendorNotFound
		}
	}
	if input.WorkOrderID != "" {
		var workOrderProperty string
		var workOrderVendor *string
		if err := tx.QueryRow(ctx, `
			SELECT mr.property_id::text,wo.vendor_id::text
			FROM work_orders wo
			JOIN maintenance_requests mr ON mr.id=wo.maintenance_request_id AND mr.organization_id=wo.organization_id
			WHERE wo.organization_id=$1 AND wo.id=$2
		`, organizationID, input.WorkOrderID).Scan(&workOrderProperty, &workOrderVendor); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return PropertyExpense{}, ErrWorkOrderNotFound
			}
			return PropertyExpense{}, err
		}
		if workOrderProperty != input.PropertyID {
			return PropertyExpense{}, ErrWorkOrderNotFound
		}
		if input.VendorID == "" && workOrderVendor != nil {
			input.VendorID = *workOrderVendor
		}
	}

	var expenseID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO property_expenses (organization_id,property_id,vendor_id,work_order_id,category,amount_minor,currency,incurred_on,note,reference_code,created_by_user_id)
		VALUES ($1,$2,NULLIF($3,'')::uuid,NULLIF($4,'')::uuid,$5,$6,$7,$8::date,$9,NULLIF($10,''),NULLIF($11,'')::uuid)
		RETURNING id
	`, organizationID, input.PropertyID, input.VendorID, input.WorkOrderID, input.Category, input.AmountMinor, input.Currency, input.IncurredOn, input.Note, input.ReferenceCode, actorUserID).Scan(&expenseID); err != nil {
		return PropertyExpense{}, err
	}
	if err := insertAudit(ctx, tx, organizationID, actorUserID, "accounting.expense.created", "property_expense", expenseID, fmt.Sprintf(`{"propertyId":"%s","amountMinor":%d,"currency":"%s"}`, input.PropertyID, input.AmountMinor, input.Currency)); err != nil {
		return PropertyExpense{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PropertyExpense{}, err
	}
	return r.getExpense(ctx, organizationID, expenseID)
}

func (r *PostgresRepository) ReverseExpense(ctx context.Context, organizationID, actorUserID string, input ReverseExpenseInput) (PropertyExpense, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return PropertyExpense{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var expenseID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM property_expenses WHERE organization_id=$1 AND id=$2 FOR UPDATE`, organizationID, input.ExpenseID).Scan(&expenseID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PropertyExpense{}, ErrExpenseNotFound
		}
		return PropertyExpense{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO property_expense_reversals (organization_id,expense_id,reason,reversed_by_user_id)
		VALUES ($1,$2,$3,NULLIF($4,'')::uuid)
	`, organizationID, input.ExpenseID, input.Reason, actorUserID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return PropertyExpense{}, ErrExpenseAlreadyReversed
		}
		return PropertyExpense{}, err
	}
	if err := insertAudit(ctx, tx, organizationID, actorUserID, "accounting.expense.reversed", "property_expense", input.ExpenseID, `{}`); err != nil {
		return PropertyExpense{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PropertyExpense{}, err
	}
	return r.getExpense(ctx, organizationID, input.ExpenseID)
}

func (r *PostgresRepository) getExpense(ctx context.Context, organizationID, expenseID string) (PropertyExpense, error) {
	var item PropertyExpense
	err := scanExpense(r.pool.QueryRow(ctx, expenseSelect+` WHERE e.organization_id=$1 AND e.id=$2`, organizationID, expenseID), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return PropertyExpense{}, ErrExpenseNotFound
	}
	return item, err
}
