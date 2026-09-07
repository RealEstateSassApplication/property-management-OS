package rent

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const obligationSelect = `
	SELECT ro.id, ro.organization_id, ro.lease_id, l.reference_code,
		p.name, u.label, primary_tenant.legal_name,
		to_char(ro.period_start, 'YYYY-MM'), to_char(ro.due_date, 'YYYY-MM-DD'),
		ro.amount_minor, alloc.allocated_minor, ro.amount_minor - alloc.allocated_minor,
		ro.currency,
		CASE
			WHEN ro.status = 'void' THEN 'void'
			WHEN alloc.allocated_minor >= ro.amount_minor THEN 'paid'
			WHEN current_date > ro.due_date THEN 'overdue'
			ELSE 'open'
		END,
		ro.created_at, ro.updated_at
	FROM rent_obligations ro
	JOIN leases l ON l.id = ro.lease_id AND l.organization_id = ro.organization_id
	JOIN tenancies tn ON tn.id = l.tenancy_id AND tn.organization_id = ro.organization_id
	JOIN units u ON u.id = tn.unit_id AND u.organization_id = ro.organization_id
	JOIN properties p ON p.id = u.property_id AND p.organization_id = ro.organization_id
	JOIN tenancy_tenants primary_link ON primary_link.tenancy_id = tn.id AND primary_link.organization_id = ro.organization_id AND primary_link.role = 'primary'
	JOIN tenants primary_tenant ON primary_tenant.id = primary_link.tenant_id AND primary_tenant.organization_id = ro.organization_id
	LEFT JOIN LATERAL (
		SELECT COALESCE(SUM(pa.amount_minor), 0)::bigint AS allocated_minor
		FROM payment_allocations pa
		WHERE pa.organization_id = ro.organization_id AND pa.obligation_id = ro.id
	) alloc ON true
`

const paymentSelect = `
	SELECT p.id, p.organization_id, p.tenant_id, t.legal_name,
		p.amount_minor, alloc.allocated_minor, p.amount_minor - alloc.allocated_minor,
		p.currency, to_char(p.received_at, 'YYYY-MM-DD'), p.method,
		COALESCE(p.reference_code, ''), p.status, p.created_at
	FROM payments p
	JOIN tenants t ON t.id = p.tenant_id AND t.organization_id = p.organization_id
	LEFT JOIN LATERAL (
		SELECT COALESCE(SUM(pa.amount_minor), 0)::bigint AS allocated_minor
		FROM payment_allocations pa
		WHERE pa.organization_id = p.organization_id AND pa.payment_id = p.id
	) alloc ON true
`

func (r *PostgresRepository) ListObligations(ctx context.Context, organizationID string) ([]Obligation, error) {
	rows, err := r.pool.Query(ctx, obligationSelect+`
		WHERE ro.organization_id = $1
		ORDER BY ro.period_start DESC, ro.due_date DESC, ro.created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Obligation, 0)
	for rows.Next() {
		var obligation Obligation
		if err := scanObligation(rows, &obligation); err != nil {
			return nil, err
		}
		result = append(result, obligation)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) getObligation(ctx context.Context, organizationID, obligationID string) (Obligation, error) {
	var obligation Obligation
	err := scanObligation(r.pool.QueryRow(ctx, obligationSelect+` WHERE ro.organization_id = $1 AND ro.id = $2`, organizationID, obligationID), &obligation)
	if errors.Is(err, pgx.ErrNoRows) {
		return Obligation{}, ErrObligationNotFound
	}
	return obligation, err
}

func (r *PostgresRepository) CreateObligation(ctx context.Context, organizationID, leaseID string, periodStart time.Time) (Obligation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Obligation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var leaseStart, leaseEnd time.Time
	var amountMinor int64
	var currency, status string
	var dueDay int
	if err := tx.QueryRow(ctx, `
		SELECT start_date, end_date, rent_amount_minor, currency, due_day, status
		FROM leases
		WHERE organization_id = $1 AND id = $2
		FOR UPDATE
	`, organizationID, leaseID).Scan(&leaseStart, &leaseEnd, &amountMinor, &currency, &dueDay, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Obligation{}, ErrLeaseNotFound
		}
		return Obligation{}, err
	}
	if status != "active" {
		return Obligation{}, ErrLeaseNotActive
	}

	period := time.Date(periodStart.Year(), periodStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	leaseStartMonth := time.Date(leaseStart.Year(), leaseStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	leaseEndMonth := time.Date(leaseEnd.Year(), leaseEnd.Month(), 1, 0, 0, 0, 0, time.UTC)
	if period.Before(leaseStartMonth) || period.After(leaseEndMonth) {
		return Obligation{}, ErrPeriodOutsideLease
	}
	lastDay := time.Date(period.Year(), period.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if dueDay > lastDay {
		dueDay = lastDay
	}
	dueDate := time.Date(period.Year(), period.Month(), dueDay, 0, 0, 0, 0, time.UTC)

	var obligationID string
	err = tx.QueryRow(ctx, `
		INSERT INTO rent_obligations (
			organization_id, lease_id, period_start, due_date, amount_minor, currency
		) VALUES ($1, $2, $3::date, $4::date, $5, $6)
		RETURNING id
	`, organizationID, leaseID, period, dueDate, amountMinor, currency).Scan(&obligationID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Obligation{}, ErrObligationExists
		}
		return Obligation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Obligation{}, err
	}
	return r.getObligation(ctx, organizationID, obligationID)
}

func (r *PostgresRepository) ListPayments(ctx context.Context, organizationID string) ([]Payment, error) {
	rows, err := r.pool.Query(ctx, paymentSelect+`
		WHERE p.organization_id = $1
		ORDER BY p.received_at DESC, p.created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Payment, 0)
	for rows.Next() {
		var payment Payment
		if err := scanPayment(rows, &payment); err != nil {
			return nil, err
		}
		result = append(result, payment)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) getPayment(ctx context.Context, organizationID, paymentID string) (Payment, error) {
	var payment Payment
	err := scanPayment(r.pool.QueryRow(ctx, paymentSelect+` WHERE p.organization_id = $1 AND p.id = $2`, organizationID, paymentID), &payment)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrPaymentNotFound
	}
	return payment, err
}

func (r *PostgresRepository) CreatePayment(ctx context.Context, organizationID string, input CreatePaymentInput) (Payment, error) {
	var tenantExists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenants WHERE organization_id = $1 AND id = $2)`, organizationID, input.TenantID).Scan(&tenantExists); err != nil {
		return Payment{}, err
	}
	if !tenantExists {
		return Payment{}, ErrTenantNotFound
	}

	var paymentID string
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO payments (
			organization_id, tenant_id, amount_minor, currency, received_at, method, reference_code
		) VALUES ($1, $2, $3, $4, $5::date, $6, NULLIF($7, ''))
		RETURNING id
	`, organizationID, input.TenantID, input.AmountMinor, input.Currency, input.ReceivedAt, input.Method, input.ReferenceCode).Scan(&paymentID); err != nil {
		return Payment{}, err
	}
	return r.getPayment(ctx, organizationID, paymentID)
}

func (r *PostgresRepository) CreateAllocation(ctx context.Context, organizationID string, input CreateAllocationInput) (Allocation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Allocation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var paymentAmount int64
	var paymentCurrency, paymentStatus, tenantID string
	if err := tx.QueryRow(ctx, `
		SELECT amount_minor, currency, status, tenant_id
		FROM payments
		WHERE organization_id = $1 AND id = $2
		FOR UPDATE
	`, organizationID, input.PaymentID).Scan(&paymentAmount, &paymentCurrency, &paymentStatus, &tenantID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Allocation{}, ErrPaymentNotFound
		}
		return Allocation{}, err
	}
	if paymentStatus != "posted" {
		return Allocation{}, ErrPaymentNotPosted
	}

	var obligationAmount int64
	var obligationCurrency, obligationStatus, tenancyID string
	if err := tx.QueryRow(ctx, `
		SELECT ro.amount_minor, ro.currency, ro.status, l.tenancy_id
		FROM rent_obligations ro
		JOIN leases l ON l.id = ro.lease_id AND l.organization_id = ro.organization_id
		WHERE ro.organization_id = $1 AND ro.id = $2
		FOR UPDATE OF ro
	`, organizationID, input.ObligationID).Scan(&obligationAmount, &obligationCurrency, &obligationStatus, &tenancyID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Allocation{}, ErrObligationNotFound
		}
		return Allocation{}, err
	}
	if obligationStatus == "void" {
		return Allocation{}, ErrObligationVoided
	}
	if paymentCurrency != obligationCurrency {
		return Allocation{}, ErrCurrencyMismatch
	}

	var tenantBelongs bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tenancy_tenants
			WHERE organization_id = $1 AND tenancy_id = $2 AND tenant_id = $3
		)
	`, organizationID, tenancyID, tenantID).Scan(&tenantBelongs); err != nil {
		return Allocation{}, err
	}
	if !tenantBelongs {
		return Allocation{}, ErrTenantMismatch
	}

	var paymentAllocated, obligationAllocated int64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount_minor), 0)::bigint FROM payment_allocations WHERE organization_id = $1 AND payment_id = $2`, organizationID, input.PaymentID).Scan(&paymentAllocated); err != nil {
		return Allocation{}, err
	}
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount_minor), 0)::bigint FROM payment_allocations WHERE organization_id = $1 AND obligation_id = $2`, organizationID, input.ObligationID).Scan(&obligationAllocated); err != nil {
		return Allocation{}, err
	}
	if input.AmountMinor > paymentAmount-paymentAllocated {
		return Allocation{}, ErrAllocationExceedsPayment
	}
	if input.AmountMinor > obligationAmount-obligationAllocated {
		return Allocation{}, ErrAllocationExceedsObligation
	}

	var allocation Allocation
	if err := tx.QueryRow(ctx, `
		INSERT INTO payment_allocations (organization_id, payment_id, obligation_id, amount_minor)
		VALUES ($1, $2, $3, $4)
		RETURNING id, organization_id, payment_id, obligation_id, amount_minor, created_at
	`, organizationID, input.PaymentID, input.ObligationID, input.AmountMinor).Scan(
		&allocation.ID, &allocation.OrganizationID, &allocation.PaymentID, &allocation.ObligationID, &allocation.AmountMinor, &allocation.CreatedAt,
	); err != nil {
		return Allocation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Allocation{}, err
	}
	return allocation, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanObligation(row scanner, obligation *Obligation) error {
	return row.Scan(
		&obligation.ID, &obligation.OrganizationID, &obligation.LeaseID, &obligation.LeaseReference,
		&obligation.PropertyName, &obligation.UnitLabel, &obligation.PrimaryTenantName,
		&obligation.Period, &obligation.DueDate, &obligation.AmountMinor, &obligation.AllocatedMinor,
		&obligation.BalanceMinor, &obligation.Currency, &obligation.State, &obligation.CreatedAt, &obligation.UpdatedAt,
	)
}

func scanPayment(row scanner, payment *Payment) error {
	return row.Scan(
		&payment.ID, &payment.OrganizationID, &payment.TenantID, &payment.TenantName,
		&payment.AmountMinor, &payment.AllocatedMinor, &payment.UnallocatedMinor, &payment.Currency,
		&payment.ReceivedAt, &payment.Method, &payment.ReferenceCode, &payment.Status, &payment.CreatedAt,
	)
}
