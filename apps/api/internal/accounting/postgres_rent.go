package accounting

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

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
	defer func() { _ = tx.Rollback(ctx) }()

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
	defer func() { _ = tx.Rollback(ctx) }()

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
