package units

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListByProperty(ctx context.Context, organizationID, propertyID string) ([]Unit, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, organization_id, property_id, reference_code, label,
		       bedrooms::float8, bathrooms::float8, floor_area::float8, floor_area_unit,
		       occupancy_status, created_at, updated_at
		FROM units
		WHERE organization_id = $1 AND property_id = $2
		ORDER BY label`, organizationID, propertyID)
	if err != nil {
		return nil, fmt.Errorf("list units: %w", err)
	}
	defer rows.Close()

	result := make([]Unit, 0)
	for rows.Next() {
		unit, err := scanUnit(rows)
		if err != nil {
			return nil, fmt.Errorf("scan unit: %w", err)
		}
		result = append(result, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate units: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, id string) (Unit, error) {
	unit, err := scanUnit(r.pool.QueryRow(ctx, `
		SELECT id, organization_id, property_id, reference_code, label,
		       bedrooms::float8, bathrooms::float8, floor_area::float8, floor_area_unit,
		       occupancy_status, created_at, updated_at
		FROM units
		WHERE organization_id = $1 AND id = $2`, organizationID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Unit{}, ErrNotFound
	}
	if err != nil {
		return Unit{}, fmt.Errorf("get unit: %w", err)
	}
	return unit, nil
}

func (r *PostgresRepository) Create(ctx context.Context, organizationID, propertyID string, input CreateInput) (Unit, error) {
	unit, err := scanUnit(r.pool.QueryRow(ctx, `
		INSERT INTO units (
			organization_id, property_id, reference_code, label,
			bedrooms, bathrooms, floor_area, floor_area_unit
		)
		SELECT $1, p.id, $3, $4, $5, $6, $7, $8
		FROM properties p
		WHERE p.id = $2 AND p.organization_id = $1
		RETURNING id, organization_id, property_id, reference_code, label,
		          bedrooms::float8, bathrooms::float8, floor_area::float8, floor_area_unit,
		          occupancy_status, created_at, updated_at`,
		organizationID, propertyID, input.ReferenceCode, input.Label, input.Bedrooms,
		input.Bathrooms, input.FloorArea, input.FloorAreaUnit,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Unit{}, ErrNotFound
	}
	if err != nil {
		return Unit{}, fmt.Errorf("create unit: %w", err)
	}
	return unit, nil
}

func (r *PostgresRepository) Update(ctx context.Context, organizationID, id string, input UpdateInput) (Unit, error) {
	unit, err := scanUnit(r.pool.QueryRow(ctx, `
		UPDATE units SET
			reference_code = COALESCE($3, reference_code),
			label = COALESCE($4, label),
			bedrooms = COALESCE($5, bedrooms),
			bathrooms = COALESCE($6, bathrooms),
			floor_area = COALESCE($7, floor_area),
			floor_area_unit = COALESCE($8, floor_area_unit),
			occupancy_status = COALESCE($9, occupancy_status),
			updated_at = now()
		WHERE organization_id = $1 AND id = $2
		RETURNING id, organization_id, property_id, reference_code, label,
		          bedrooms::float8, bathrooms::float8, floor_area::float8, floor_area_unit,
		          occupancy_status, created_at, updated_at`,
		organizationID, id, input.ReferenceCode, input.Label, input.Bedrooms,
		input.Bathrooms, input.FloorArea, input.FloorAreaUnit, input.OccupancyStatus,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Unit{}, ErrNotFound
	}
	if err != nil {
		return Unit{}, fmt.Errorf("update unit: %w", err)
	}
	return unit, nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanUnit(row rowScanner) (Unit, error) {
	var unit Unit
	err := row.Scan(
		&unit.ID, &unit.OrganizationID, &unit.PropertyID, &unit.ReferenceCode, &unit.Label,
		&unit.Bedrooms, &unit.Bathrooms, &unit.FloorArea, &unit.FloorAreaUnit,
		&unit.OccupancyStatus, &unit.CreatedAt, &unit.UpdatedAt,
	)
	return unit, err
}
