package tenancies

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

const tenancySelect = `
	SELECT t.id, t.organization_id, t.unit_id, u.label, p.name,
		primary_link.tenant_id, primary_tenant.legal_name,
		GREATEST(COUNT(all_links.tenant_id) - 1, 0)::int,
		to_char(t.start_date, 'YYYY-MM-DD'), COALESCE(to_char(t.end_date, 'YYYY-MM-DD'), ''),
		t.status, t.created_at, t.updated_at
	FROM tenancies t
	JOIN units u ON u.id = t.unit_id AND u.organization_id = t.organization_id
	JOIN properties p ON p.id = u.property_id AND p.organization_id = t.organization_id
	JOIN tenancy_tenants primary_link ON primary_link.tenancy_id = t.id AND primary_link.organization_id = t.organization_id AND primary_link.role = 'primary'
	JOIN tenants primary_tenant ON primary_tenant.id = primary_link.tenant_id AND primary_tenant.organization_id = t.organization_id
	LEFT JOIN tenancy_tenants all_links ON all_links.tenancy_id = t.id AND all_links.organization_id = t.organization_id
`

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]Tenancy, error) {
	rows, err := r.pool.Query(ctx, tenancySelect+`
		WHERE t.organization_id = $1
		GROUP BY t.id, u.label, p.name, primary_link.tenant_id, primary_tenant.legal_name
		ORDER BY t.start_date DESC, t.created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Tenancy, 0)
	for rows.Next() {
		var tenancy Tenancy
		if err := scanTenancy(rows, &tenancy); err != nil {
			return nil, err
		}
		result = append(result, tenancy)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, tenancyID string) (Tenancy, error) {
	row := r.pool.QueryRow(ctx, tenancySelect+`
		WHERE t.organization_id = $1 AND t.id = $2
		GROUP BY t.id, u.label, p.name, primary_link.tenant_id, primary_tenant.legal_name
	`, organizationID, tenancyID)
	var tenancy Tenancy
	if err := scanTenancy(row, &tenancy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tenancy{}, ErrNotFound
		}
		return Tenancy{}, err
	}
	return tenancy, nil
}

func (r *PostgresRepository) Create(ctx context.Context, organizationID string, input CreateInput) (Tenancy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Tenancy{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var unitID string
	if err := tx.QueryRow(ctx, `SELECT id FROM units WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, input.UnitID).Scan(&unitID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tenancy{}, ErrUnitNotFound
		}
		return Tenancy{}, err
	}
	if input.Status == "active" {
		var existing string
		err := tx.QueryRow(ctx, `SELECT id FROM tenancies WHERE organization_id = $1 AND unit_id = $2 AND status = 'active' LIMIT 1`, organizationID, input.UnitID).Scan(&existing)
		if err == nil {
			return Tenancy{}, ErrUnitConflict
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Tenancy{}, err
		}
	}

	tenantIDs := append([]string{input.PrimaryTenantID}, input.OccupantTenantIDs...)
	for _, tenantID := range tenantIDs {
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM tenants WHERE organization_id = $1 AND id = $2`, organizationID, tenantID).Scan(&status); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Tenancy{}, ErrTenantNotFound
			}
			return Tenancy{}, err
		}
		if status == "blocked" {
			return Tenancy{}, ErrTenantInvalid
		}
	}

	var tenancyID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO tenancies (organization_id, unit_id, start_date, end_date, status)
		VALUES ($1, $2, $3::date, NULLIF($4, '')::date, $5)
		RETURNING id
	`, organizationID, input.UnitID, input.StartDate, input.EndDate, input.Status).Scan(&tenancyID); err != nil {
		return Tenancy{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO tenancy_tenants (organization_id, tenancy_id, tenant_id, role)
		VALUES ($1, $2, $3, 'primary')
	`, organizationID, tenancyID, input.PrimaryTenantID); err != nil {
		return Tenancy{}, err
	}
	for _, tenantID := range input.OccupantTenantIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO tenancy_tenants (organization_id, tenancy_id, tenant_id, role)
			VALUES ($1, $2, $3, 'occupant')
		`, organizationID, tenancyID, tenantID); err != nil {
			return Tenancy{}, err
		}
	}

	if input.Status == "active" {
		if _, err := tx.Exec(ctx, `UPDATE units SET occupancy_status = 'occupied', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, input.UnitID); err != nil {
			return Tenancy{}, err
		}
		for _, tenantID := range tenantIDs {
			if _, err := tx.Exec(ctx, `UPDATE tenants SET status = 'active', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, tenantID); err != nil {
				return Tenancy{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Tenancy{}, err
	}
	return r.Get(ctx, organizationID, tenancyID)
}

func (r *PostgresRepository) Update(ctx context.Context, organizationID, tenancyID string, input UpdateInput) (Tenancy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Tenancy{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var unitID, currentStatus string
	if err := tx.QueryRow(ctx, `SELECT unit_id, status FROM tenancies WHERE organization_id = $1 AND id = $2 FOR UPDATE`, organizationID, tenancyID).Scan(&unitID, &currentStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tenancy{}, ErrNotFound
		}
		return Tenancy{}, err
	}

	endDateSet := input.EndDate != nil
	statusSet := input.Status != nil
	endDate, status := "", currentStatus
	if input.EndDate != nil {
		endDate = *input.EndDate
	}
	if input.Status != nil {
		status = *input.Status
	}
	if status == "active" && currentStatus != "active" {
		var existing string
		err := tx.QueryRow(ctx, `SELECT id FROM tenancies WHERE organization_id = $1 AND unit_id = $2 AND status = 'active' AND id <> $3 LIMIT 1`, organizationID, unitID, tenancyID).Scan(&existing)
		if err == nil {
			return Tenancy{}, ErrUnitConflict
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Tenancy{}, err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE tenancies
		SET end_date = CASE WHEN $3 THEN NULLIF($4, '')::date ELSE end_date END,
			status = CASE WHEN $5 THEN $6 ELSE status END,
			updated_at = now()
		WHERE organization_id = $1 AND id = $2
	`, organizationID, tenancyID, endDateSet, endDate, statusSet, status); err != nil {
		return Tenancy{}, err
	}

	if status == "active" {
		if _, err := tx.Exec(ctx, `UPDATE units SET occupancy_status = 'occupied', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, unitID); err != nil {
			return Tenancy{}, err
		}
	} else if currentStatus == "active" && (status == "ended" || status == "cancelled") {
		if _, err := tx.Exec(ctx, `UPDATE units SET occupancy_status = 'vacant', updated_at = now() WHERE organization_id = $1 AND id = $2`, organizationID, unitID); err != nil {
			return Tenancy{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Tenancy{}, err
	}
	return r.Get(ctx, organizationID, tenancyID)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTenancy(row scanner, tenancy *Tenancy) error {
	return row.Scan(
		&tenancy.ID, &tenancy.OrganizationID, &tenancy.UnitID, &tenancy.UnitLabel, &tenancy.PropertyName,
		&tenancy.PrimaryTenantID, &tenancy.PrimaryTenantName, &tenancy.OccupantCount,
		&tenancy.StartDate, &tenancy.EndDate, &tenancy.Status, &tenancy.CreatedAt, &tenancy.UpdatedAt,
	)
}
