package leases

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const leaseSelect = `
	SELECT l.id, l.organization_id, l.tenancy_id, l.reference_code,
		p.name, u.label, primary_tenant.legal_name,
		to_char(l.start_date, 'YYYY-MM-DD'), to_char(l.end_date, 'YYYY-MM-DD'),
		l.rent_amount_minor, l.deposit_amount_minor, l.currency, l.due_day,
		l.status, l.created_at, l.updated_at
	FROM leases l
	JOIN tenancies t ON t.id = l.tenancy_id AND t.organization_id = l.organization_id
	JOIN units u ON u.id = t.unit_id AND u.organization_id = l.organization_id
	JOIN properties p ON p.id = u.property_id AND p.organization_id = l.organization_id
	JOIN tenancy_tenants primary_link ON primary_link.tenancy_id = t.id AND primary_link.organization_id = l.organization_id AND primary_link.role = 'primary'
	JOIN tenants primary_tenant ON primary_tenant.id = primary_link.tenant_id AND primary_tenant.organization_id = l.organization_id
`

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]Lease, error) {
	rows, err := r.pool.Query(ctx, leaseSelect+`
		WHERE l.organization_id = $1
		ORDER BY l.start_date DESC, l.created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Lease, 0)
	for rows.Next() {
		var lease Lease
		if err := scanLease(rows, &lease); err != nil {
			return nil, err
		}
		result = append(result, lease)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, leaseID string) (Lease, error) {
	row := r.pool.QueryRow(ctx, leaseSelect+` WHERE l.organization_id = $1 AND l.id = $2`, organizationID, leaseID)
	var lease Lease
	if err := scanLease(row, &lease); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lease{}, ErrNotFound
		}
		return Lease{}, err
	}
	return lease, nil
}

func (r *PostgresRepository) Create(ctx context.Context, organizationID string, input CreateInput) (Lease, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Lease{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var unitID, tenancyStatus string
	if err := tx.QueryRow(ctx, `
		SELECT unit_id, status FROM tenancies
		WHERE organization_id = $1 AND id = $2
		FOR UPDATE
	`, organizationID, input.TenancyID).Scan(&unitID, &tenancyStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lease{}, ErrTenancyNotFound
		}
		return Lease{}, err
	}
	if input.Status == "active" {
		var existing string
		err := tx.QueryRow(ctx, `SELECT id FROM leases WHERE organization_id = $1 AND tenancy_id = $2 AND status = 'active' LIMIT 1`, organizationID, input.TenancyID).Scan(&existing)
		if err == nil {
			return Lease{}, ErrActiveLeaseConflict
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Lease{}, err
		}
	}

	var leaseID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO leases (
			organization_id, tenancy_id, reference_code, start_date, end_date,
			rent_amount_minor, deposit_amount_minor, currency, due_day, status
		) VALUES ($1, $2, $3, $4::date, $5::date, $6, $7, $8, $9, $10)
		RETURNING id
	`, organizationID, input.TenancyID, input.ReferenceCode, input.StartDate, input.EndDate,
		input.RentAmountMinor, input.DepositAmountMinor, input.Currency, input.DueDay, input.Status).Scan(&leaseID); err != nil {
		return Lease{}, err
	}

	if input.Status == "active" {
		if tenancyStatus == "upcoming" {
			if _, err := tx.Exec(ctx, `UPDATE tenancies SET status = 'active', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, input.TenancyID); err != nil {
				return Lease{}, err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE units SET occupancy_status = 'occupied', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, unitID); err != nil {
			return Lease{}, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE tenants SET status = 'active', updated_at = now()
			WHERE organization_id = $1 AND id IN (
				SELECT tenant_id FROM tenancy_tenants WHERE organization_id = $1 AND tenancy_id = $2
			)
		`, organizationID, input.TenancyID); err != nil {
			return Lease{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Lease{}, err
	}
	return r.Get(ctx, organizationID, leaseID)
}

func (r *PostgresRepository) Update(ctx context.Context, organizationID, leaseID string, input UpdateInput) (Lease, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Lease{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var tenancyID, unitID, currentStatus string
	if err := tx.QueryRow(ctx, `
		SELECT l.tenancy_id, t.unit_id, l.status
		FROM leases l
		JOIN tenancies t ON t.id = l.tenancy_id AND t.organization_id = l.organization_id
		WHERE l.organization_id = $1 AND l.id = $2
		FOR UPDATE OF l, t
	`, organizationID, leaseID).Scan(&tenancyID, &unitID, &currentStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lease{}, ErrNotFound
		}
		return Lease{}, err
	}

	status := currentStatus
	if input.Status != nil {
		status = *input.Status
	}
	if status == "active" && currentStatus != "active" {
		var existing string
		err := tx.QueryRow(ctx, `SELECT id FROM leases WHERE organization_id = $1 AND tenancy_id = $2 AND status = 'active' AND id <> $3 LIMIT 1`, organizationID, tenancyID, leaseID).Scan(&existing)
		if err == nil {
			return Lease{}, ErrActiveLeaseConflict
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Lease{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE leases SET status = $3, updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, leaseID, status); err != nil {
		return Lease{}, err
	}

	if status == "active" {
		if _, err := tx.Exec(ctx, `UPDATE tenancies SET status = 'active', updated_at = now() WHERE organization_id = $1 AND id = $2 AND status = 'upcoming'`, organizationID, tenancyID); err != nil {
			return Lease{}, err
		}
		if _, err := tx.Exec(ctx, `UPDATE units SET occupancy_status = 'occupied', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, unitID); err != nil {
			return Lease{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Lease{}, err
	}
	return r.Get(ctx, organizationID, leaseID)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanLease(row scanner, lease *Lease) error {
	return row.Scan(
		&lease.ID, &lease.OrganizationID, &lease.TenancyID, &lease.ReferenceCode,
		&lease.PropertyName, &lease.UnitLabel, &lease.PrimaryTenantName,
		&lease.StartDate, &lease.EndDate, &lease.RentAmountMinor, &lease.DepositAmountMinor,
		&lease.Currency, &lease.DueDay, &lease.Status, &lease.CreatedAt, &lease.UpdatedAt,
	)
}
