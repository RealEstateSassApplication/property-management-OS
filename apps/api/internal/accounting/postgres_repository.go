package accounting

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const adjustmentSelect = `
	SELECT ra.id,ra.organization_id,ra.obligation_id,l.reference_code,t.legal_name,
	       ra.adjustment_type,ra.amount_minor,ro.currency,ra.reason,
	       COALESCE(ra.created_by_user_id::text,''),ra.created_at
	FROM rent_adjustments ra
	JOIN rent_obligations ro ON ro.id=ra.obligation_id AND ro.organization_id=ra.organization_id
	JOIN leases l ON l.id=ro.lease_id AND l.organization_id=ro.organization_id
	JOIN tenancies tn ON tn.id=l.tenancy_id AND tn.organization_id=l.organization_id
	JOIN tenancy_tenants tt ON tt.tenancy_id=tn.id AND tt.organization_id=tn.organization_id AND tt.role='primary'
	JOIN tenants t ON t.id=tt.tenant_id AND t.organization_id=tt.organization_id
`

func (r *PostgresRepository) ListRentAdjustments(ctx context.Context, organizationID string) ([]RentAdjustment, error) {
	rows, err := r.pool.Query(ctx, adjustmentSelect+` WHERE ra.organization_id=$1 ORDER BY ra.created_at DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]RentAdjustment, 0)
	for rows.Next() {
		var item RentAdjustment
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.ObligationID, &item.LeaseReference, &item.TenantName, &item.AdjustmentType, &item.AmountMinor, &item.Currency, &item.Reason, &item.CreatedByUserID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateRentAdjustment(ctx context.Context, organizationID, actorUserID string, input CreateRentAdjustmentInput) (RentAdjustment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return RentAdjustment{}, err
	}
	defer tx.Rollback(ctx)

	var currentAmount, allocated int64
	var status string
	if err := tx.QueryRow(ctx, `
		SELECT ro.amount_minor,ro.status,
		       COALESCE((SELECT SUM(pa.amount_minor) FROM payment_allocations pa WHERE pa.organization_id=ro.organization_id AND pa.obligation_id=ro.id),0)::bigint
		FROM rent_obligations ro
		WHERE ro.organization_id=$1 AND ro.id=$2
		FOR UPDATE
	`, organizationID, input.ObligationID).Scan(&currentAmount, &status, &allocated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RentAdjustment{}, ErrObligationNotFound
		}
		return RentAdjustment{}, err
	}
	if status == "void" {
		return RentAdjustment{}, ErrObligationNotFound
	}
	if input.AdjustmentType == "credit" || input.AdjustmentType == "writeoff" {
		if currentAmount-input.AmountMinor < allocated || currentAmount-input.AmountMinor <= 0 {
			return RentAdjustment{}, ErrAdjustmentWouldOverpay
		}
	}

	var adjustmentID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO rent_adjustments (organization_id,obligation_id,adjustment_type,amount_minor,reason,created_by_user_id)
		VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::uuid)
		RETURNING id
	`, organizationID, input.ObligationID, input.AdjustmentType, input.AmountMinor, input.Reason, actorUserID).Scan(&adjustmentID); err != nil {
		return RentAdjustment{}, err
	}
	if err := insertAudit(ctx, tx, organizationID, actorUserID, "accounting.rent_adjustment.created", "rent_obligation", input.ObligationID, fmt.Sprintf(`{"adjustmentId":"%s","type":"%s","amountMinor":%d}`, adjustmentID, input.AdjustmentType, input.AmountMinor)); err != nil {
		return RentAdjustment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RentAdjustment{}, err
	}
	return r.getRentAdjustment(ctx, organizationID, adjustmentID)
}

func (r *PostgresRepository) getRentAdjustment(ctx context.Context, organizationID, adjustmentID string) (RentAdjustment, error) {
	var item RentAdjustment
	err := r.pool.QueryRow(ctx, adjustmentSelect+` WHERE ra.organization_id=$1 AND ra.id=$2`, organizationID, adjustmentID).Scan(
		&item.ID, &item.OrganizationID, &item.ObligationID, &item.LeaseReference, &item.TenantName,
		&item.AdjustmentType, &item.AmountMinor, &item.Currency, &item.Reason, &item.CreatedByUserID, &item.CreatedAt,
	)
	return item, err
}

const reversalSelect = `
	SELECT pr.id,pr.organization_id,pr.payment_id,t.legal_name,p.amount_minor,p.currency,
	       COALESCE(p.reference_code,''),pr.reason,COALESCE(pr.reversed_by_user_id::text,''),pr.reversed_at
	FROM payment_reversals pr
	JOIN payments p ON p.id=pr.payment_id AND p.organization_id=pr.organization_id
	JOIN tenants t ON t.id=p.tenant_id AND t.organization_id=p.organization_id
`

func (r *PostgresRepository) ListPaymentReversals(ctx context.Context, organizationID string) ([]PaymentReversal, error) {
	rows, err := r.pool.Query(ctx, reversalSelect+` WHERE pr.organization_id=$1 ORDER BY pr.reversed_at DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PaymentReversal, 0)
	for rows.Next() {
		var item PaymentReversal
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.PaymentID, &item.TenantName, &item.AmountMinor, &item.Currency, &item.ReferenceCode, &item.Reason, &item.ReversedByUserID, &item.ReversedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ReversePayment(ctx context.Context, organizationID, actorUserID string, input ReversePaymentInput) (PaymentReversal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return PaymentReversal{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM payments WHERE organization_id=$1 AND id=$2 FOR UPDATE`, organizationID, input.PaymentID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PaymentReversal{}, ErrPaymentNotFound
		}
		return PaymentReversal{}, err
	}
	if status != "posted" {
		var already bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM payment_reversals WHERE organization_id=$1 AND payment_id=$2)`, organizationID, input.PaymentID).Scan(&already); err != nil {
			return PaymentReversal{}, err
		}
		if already {
			return PaymentReversal{}, ErrPaymentAlreadyReversed
		}
		return PaymentReversal{}, ErrPaymentNotPosted
	}

	var reversalID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO payment_reversals (organization_id,payment_id,reason,reversed_by_user_id)
		VALUES ($1,$2,$3,NULLIF($4,'')::uuid)
		RETURNING id
	`, organizationID, input.PaymentID, input.Reason, actorUserID).Scan(&reversalID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return PaymentReversal{}, ErrPaymentAlreadyReversed
		}
		return PaymentReversal{}, err
	}

	rows, err := tx.Query(ctx, `
		SELECT obligation_id,COALESCE(SUM(amount_minor),0)::bigint
		FROM payment_allocations
		WHERE organization_id=$1 AND payment_id=$2
		GROUP BY obligation_id
	`, organizationID, input.PaymentID)
	if err != nil {
		return PaymentReversal{}, err
	}
	type allocationTotal struct {
		obligationID string
		amountMinor  int64
	}
	allocations := make([]allocationTotal, 0)
	for rows.Next() {
		var item allocationTotal
		if err := rows.Scan(&item.obligationID, &item.amountMinor); err != nil {
			rows.Close()
			return PaymentReversal{}, err
		}
		allocations = append(allocations, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return PaymentReversal{}, err
	}
	rows.Close()

	for _, allocation := range allocations {
		if _, err := tx.Exec(ctx, `
			INSERT INTO rent_adjustments (organization_id,obligation_id,adjustment_type,amount_minor,reason,created_by_user_id)
			VALUES ($1,$2,'payment_reversal',$3,$4,NULLIF($5,'')::uuid)
		`, organizationID, allocation.obligationID, allocation.amountMinor, "Payment reversal: "+input.Reason, actorUserID); err != nil {
			return PaymentReversal{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE payments SET status='void' WHERE organization_id=$1 AND id=$2`, organizationID, input.PaymentID); err != nil {
		return PaymentReversal{}, err
	}
	if err := insertAudit(ctx, tx, organizationID, actorUserID, "accounting.payment.reversed", "payment", input.PaymentID, fmt.Sprintf(`{"reversalId":"%s"}`, reversalID)); err != nil {
		return PaymentReversal{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PaymentReversal{}, err
	}
	return r.getPaymentReversal(ctx, organizationID, reversalID)
}

func (r *PostgresRepository) getPaymentReversal(ctx context.Context, organizationID, reversalID string) (PaymentReversal, error) {
	var item PaymentReversal
	err := r.pool.QueryRow(ctx, reversalSelect+` WHERE pr.organization_id=$1 AND pr.id=$2`, organizationID, reversalID).Scan(
		&item.ID, &item.OrganizationID, &item.PaymentID, &item.TenantName, &item.AmountMinor, &item.Currency,
		&item.ReferenceCode, &item.Reason, &item.ReversedByUserID, &item.ReversedAt,
	)
	return item, err
}

const depositAccountSelect = `
	SELECT da.id,da.organization_id,da.lease_id,l.reference_code,t.legal_name,p.name,u.label,
	       da.required_amount_minor,
	       COALESCE(SUM(CASE WHEN dt.transaction_type IN ('received','adjustment_increase') THEN dt.amount_minor ELSE -dt.amount_minor END),0)::bigint,
	       da.currency,da.created_at
	FROM security_deposit_accounts da
	JOIN leases l ON l.id=da.lease_id AND l.organization_id=da.organization_id
	JOIN tenancies tn ON tn.id=l.tenancy_id AND tn.organization_id=l.organization_id
	JOIN units u ON u.id=tn.unit_id AND u.organization_id=tn.organization_id
	JOIN properties p ON p.id=u.property_id AND p.organization_id=u.organization_id
	JOIN tenancy_tenants tt ON tt.tenancy_id=tn.id AND tt.organization_id=tn.organization_id AND tt.role='primary'
	JOIN tenants t ON t.id=tt.tenant_id AND t.organization_id=tt.organization_id
	LEFT JOIN security_deposit_transactions dt ON dt.deposit_account_id=da.id AND dt.organization_id=da.organization_id
`

func (r *PostgresRepository) ListDepositAccounts(ctx context.Context, organizationID string) ([]DepositAccount, error) {
	rows, err := r.pool.Query(ctx, depositAccountSelect+`
		WHERE da.organization_id=$1
		GROUP BY da.id,l.reference_code,t.legal_name,p.name,u.label
		ORDER BY da.created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]DepositAccount, 0)
	for rows.Next() {
		var item DepositAccount
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.LeaseID, &item.LeaseReference, &item.TenantName, &item.PropertyName, &item.UnitLabel, &item.RequiredAmountMinor, &item.HeldAmountMinor, &item.Currency, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateDepositAccount(ctx context.Context, organizationID string, input CreateDepositAccountInput) (DepositAccount, error) {
	var required int64
	var currency string
	if err := r.pool.QueryRow(ctx, `SELECT deposit_amount_minor,currency FROM leases WHERE organization_id=$1 AND id=$2`, organizationID, input.LeaseID).Scan(&required, &currency); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DepositAccount{}, ErrLeaseNotFound
		}
		return DepositAccount{}, err
	}
	var accountID string
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO security_deposit_accounts (organization_id,lease_id,required_amount_minor,currency)
		VALUES ($1,$2,$3,$4)
		RETURNING id
	`, organizationID, input.LeaseID, required, currency).Scan(&accountID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return DepositAccount{}, ErrDepositAccountExists
		}
		return DepositAccount{}, err
	}
	return r.getDepositAccount(ctx, organizationID, accountID)
}

func (r *PostgresRepository) getDepositAccount(ctx context.Context, organizationID, accountID string) (DepositAccount, error) {
	var item DepositAccount
	err := r.pool.QueryRow(ctx, depositAccountSelect+`
		WHERE da.organization_id=$1 AND da.id=$2
		GROUP BY da.id,l.reference_code,t.legal_name,p.name,u.label
	`, organizationID, accountID).Scan(
		&item.ID, &item.OrganizationID, &item.LeaseID, &item.LeaseReference, &item.TenantName, &item.PropertyName,
		&item.UnitLabel, &item.RequiredAmountMinor, &item.HeldAmountMinor, &item.Currency, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DepositAccount{}, ErrDepositAccountNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListDepositTransactions(ctx context.Context, organizationID string) ([]DepositTransaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,organization_id,deposit_account_id,transaction_type,amount_minor,to_char(occurred_on,'YYYY-MM-DD'),note,
		       COALESCE(created_by_user_id::text,''),created_at
		FROM security_deposit_transactions
		WHERE organization_id=$1
		ORDER BY occurred_on DESC,created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]DepositTransaction, 0)
	for rows.Next() {
		var item DepositTransaction
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.DepositAccountID, &item.TransactionType, &item.AmountMinor, &item.OccurredOn, &item.Note, &item.CreatedByUserID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateDepositTransaction(ctx context.Context, organizationID, actorUserID string, input CreateDepositTransactionInput) (DepositTransaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return DepositTransaction{}, err
	}
	defer tx.Rollback(ctx)
	var held int64
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE WHEN dt.transaction_type IN ('received','adjustment_increase') THEN dt.amount_minor ELSE -dt.amount_minor END),0)::bigint
		FROM security_deposit_accounts da
		LEFT JOIN security_deposit_transactions dt ON dt.deposit_account_id=da.id AND dt.organization_id=da.organization_id
		WHERE da.organization_id=$1 AND da.id=$2
		GROUP BY da.id
		FOR UPDATE OF da
	`, organizationID, input.DepositAccountID).Scan(&held); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DepositTransaction{}, ErrDepositAccountNotFound
		}
		return DepositTransaction{}, err
	}
	if input.TransactionType == "deduction" || input.TransactionType == "refund" || input.TransactionType == "adjustment_decrease" {
		if input.AmountMinor > held {
			return DepositTransaction{}, ErrDepositInsufficientFunds
		}
	}
	var item DepositTransaction
	if err := tx.QueryRow(ctx, `
		INSERT INTO security_deposit_transactions (organization_id,deposit_account_id,transaction_type,amount_minor,occurred_on,note,created_by_user_id)
		VALUES ($1,$2,$3,$4,$5::date,$6,NULLIF($7,'')::uuid)
		RETURNING id,organization_id,deposit_account_id,transaction_type,amount_minor,to_char(occurred_on,'YYYY-MM-DD'),note,COALESCE(created_by_user_id::text,''),created_at
	`, organizationID, input.DepositAccountID, input.TransactionType, input.AmountMinor, input.OccurredOn, input.Note, actorUserID).Scan(
		&item.ID, &item.OrganizationID, &item.DepositAccountID, &item.TransactionType, &item.AmountMinor, &item.OccurredOn, &item.Note, &item.CreatedByUserID, &item.CreatedAt,
	); err != nil {
		return DepositTransaction{}, err
	}
	if err := insertAudit(ctx, tx, organizationID, actorUserID, "accounting.deposit.transaction_created", "security_deposit_account", input.DepositAccountID, fmt.Sprintf(`{"transactionId":"%s","type":"%s","amountMinor":%d}`, item.ID, item.TransactionType, item.AmountMinor)); err != nil {
		return DepositTransaction{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DepositTransaction{}, err
	}
	return item, nil
}

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
	defer tx.Rollback(ctx)
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
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM property_expenses WHERE organization_id=$1 AND id=$2)`, organizationID, input.ExpenseID).Scan(&exists); err != nil {
		return PropertyExpense{}, err
	}
	if !exists {
		return PropertyExpense{}, ErrExpenseNotFound
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

func (r *PostgresRepository) OwnerStatement(ctx context.Context, organizationID, ownerID string, from, to time.Time) (OwnerStatement, error) {
	var statement OwnerStatement
	statement.OwnerID = ownerID
	statement.From = from.Format("2006-01-02")
	statement.To = to.Format("2006-01-02")
	if err := r.pool.QueryRow(ctx, `SELECT legal_name FROM owners WHERE organization_id=$1 AND id=$2 AND status='active'`, organizationID, ownerID).Scan(&statement.OwnerName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OwnerStatement{}, ErrOwnerNotFound
		}
		return OwnerStatement{}, err
	}
	lines := make([]StatementLine, 0)
	incomeRows, err := r.pool.Query(ctx, `
		SELECT to_char(pmt.received_at,'YYYY-MM-DD'),pr.id,pr.name,
		       'Rent collection',pmt.currency,pa.amount_minor,oi.ownership_bps,
		       (pa.amount_minor * oi.ownership_bps / 10000)::bigint
		FROM payment_allocations pa
		JOIN payments pmt ON pmt.id=pa.payment_id AND pmt.organization_id=pa.organization_id AND pmt.status='posted'
		JOIN rent_obligations ro ON ro.id=pa.obligation_id AND ro.organization_id=pa.organization_id
		JOIN leases l ON l.id=ro.lease_id AND l.organization_id=ro.organization_id
		JOIN tenancies tn ON tn.id=l.tenancy_id AND tn.organization_id=l.organization_id
		JOIN units u ON u.id=tn.unit_id AND u.organization_id=tn.organization_id
		JOIN properties pr ON pr.id=u.property_id AND pr.organization_id=u.organization_id
		JOIN ownership_interests oi ON oi.organization_id=pr.organization_id AND oi.property_id=pr.id AND oi.owner_id=$2
		WHERE pa.organization_id=$1 AND pmt.received_at BETWEEN $3::date AND $4::date
		  AND oi.effective_from<=pmt.received_at
		  AND (oi.effective_to IS NULL OR oi.effective_to>=pmt.received_at)
		ORDER BY pmt.received_at,pa.created_at
	`, organizationID, ownerID, from, to)
	if err != nil {
		return OwnerStatement{}, err
	}
	for incomeRows.Next() {
		var line StatementLine
		line.LineType = "income"
		if err := incomeRows.Scan(&line.Date, &line.PropertyID, &line.PropertyName, &line.Description, &line.Currency, &line.GrossMinor, &line.OwnershipBPS, &line.OwnerMinor); err != nil {
			incomeRows.Close()
			return OwnerStatement{}, err
		}
		lines = append(lines, line)
	}
	if err := incomeRows.Err(); err != nil {
		incomeRows.Close()
		return OwnerStatement{}, err
	}
	incomeRows.Close()

	expenseRows, err := r.pool.Query(ctx, `
		SELECT to_char(e.incurred_on,'YYYY-MM-DD'),pr.id,pr.name,
		       e.category || ': ' || e.note,e.currency,e.amount_minor,oi.ownership_bps,
		       (e.amount_minor * oi.ownership_bps / 10000)::bigint
		FROM property_expenses e
		JOIN properties pr ON pr.id=e.property_id AND pr.organization_id=e.organization_id
		JOIN ownership_interests oi ON oi.organization_id=pr.organization_id AND oi.property_id=pr.id AND oi.owner_id=$2
		LEFT JOIN property_expense_reversals er ON er.organization_id=e.organization_id AND er.expense_id=e.id
		WHERE e.organization_id=$1 AND e.incurred_on BETWEEN $3::date AND $4::date
		  AND er.id IS NULL
		  AND oi.effective_from<=e.incurred_on
		  AND (oi.effective_to IS NULL OR oi.effective_to>=e.incurred_on)
		ORDER BY e.incurred_on,e.created_at
	`, organizationID, ownerID, from, to)
	if err != nil {
		return OwnerStatement{}, err
	}
	for expenseRows.Next() {
		var line StatementLine
		line.LineType = "expense"
		if err := expenseRows.Scan(&line.Date, &line.PropertyID, &line.PropertyName, &line.Description, &line.Currency, &line.GrossMinor, &line.OwnershipBPS, &line.OwnerMinor); err != nil {
			expenseRows.Close()
			return OwnerStatement{}, err
		}
		lines = append(lines, line)
	}
	if err := expenseRows.Err(); err != nil {
		expenseRows.Close()
		return OwnerStatement{}, err
	}
	expenseRows.Close()

	sort.SliceStable(lines, func(i, j int) bool {
		if lines[i].Date == lines[j].Date {
			return lines[i].LineType < lines[j].LineType
		}
		return lines[i].Date < lines[j].Date
	})
	statement.Lines = lines
	summaryMap := make(map[string]*StatementCurrencySummary)
	for _, line := range lines {
		summary := summaryMap[line.Currency]
		if summary == nil {
			summary = &StatementCurrencySummary{Currency: line.Currency}
			summaryMap[line.Currency] = summary
		}
		if line.LineType == "income" {
			summary.IncomeMinor += line.OwnerMinor
		} else {
			summary.ExpenseMinor += line.OwnerMinor
		}
	}
	currencies := make([]string, 0, len(summaryMap))
	for currency := range summaryMap {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	statement.Summaries = make([]StatementCurrencySummary, 0, len(currencies))
	for _, currency := range currencies {
		summary := *summaryMap[currency]
		summary.NetOwnerAmountMinor = summary.IncomeMinor - summary.ExpenseMinor
		statement.Summaries = append(statement.Summaries, summary)
	}
	return statement, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExpense(row scanner, item *PropertyExpense) error {
	return row.Scan(
		&item.ID, &item.OrganizationID, &item.PropertyID, &item.PropertyName, &item.VendorID, &item.VendorName,
		&item.WorkOrderID, &item.Category, &item.AmountMinor, &item.Currency, &item.IncurredOn, &item.Note,
		&item.ReferenceCode, &item.CreatedByUserID, &item.CreatedAt, &item.Reversed, &item.ReversalReason, &item.ReversedAt,
	)
}

func insertAudit(ctx context.Context, tx pgx.Tx, organizationID, actorUserID, action, resourceType, resourceID, metadata string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_events (organization_id,actor_user_id,action,resource_type,resource_id,metadata)
		VALUES ($1,NULLIF($2,'')::uuid,$3,$4,NULLIF($5,'')::uuid,$6::jsonb)
	`, organizationID, actorUserID, action, resourceType, resourceID, metadata)
	return err
}
