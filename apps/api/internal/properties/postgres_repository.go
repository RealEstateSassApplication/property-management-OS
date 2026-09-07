package properties

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

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]Property, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, organization_id, reference_code, name, property_type,
		       address_line_1, address_line_2, city, region, postal_code,
		       country_code, status, external_avara_property_id, created_at, updated_at
		FROM properties
		WHERE organization_id = $1
		ORDER BY created_at DESC`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list properties: %w", err)
	}
	defer rows.Close()

	result := make([]Property, 0)
	for rows.Next() {
		property, err := scanProperty(rows)
		if err != nil {
			return nil, fmt.Errorf("scan property: %w", err)
		}
		result = append(result, property)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate properties: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, id string) (Property, error) {
	property, err := scanProperty(r.pool.QueryRow(ctx, `
		SELECT id, organization_id, reference_code, name, property_type,
		       address_line_1, address_line_2, city, region, postal_code,
		       country_code, status, external_avara_property_id, created_at, updated_at
		FROM properties
		WHERE organization_id = $1 AND id = $2`, organizationID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Property{}, ErrNotFound
	}
	if err != nil {
		return Property{}, fmt.Errorf("get property: %w", err)
	}
	return property, nil
}

func (r *PostgresRepository) Create(ctx context.Context, organizationID string, input CreateInput) (Property, error) {
	property, err := scanProperty(r.pool.QueryRow(ctx, `
		INSERT INTO properties (
			organization_id, reference_code, name, property_type, address_line_1,
			address_line_2, city, region, postal_code, country_code, external_avara_property_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, organization_id, reference_code, name, property_type,
		          address_line_1, address_line_2, city, region, postal_code,
		          country_code, status, external_avara_property_id, created_at, updated_at`,
		organizationID, input.ReferenceCode, input.Name, input.PropertyType, input.AddressLine1,
		input.AddressLine2, input.City, input.Region, input.PostalCode, input.CountryCode,
		input.ExternalAvaraPropertyID,
	))
	if err != nil {
		return Property{}, fmt.Errorf("create property: %w", err)
	}
	return property, nil
}

func (r *PostgresRepository) Update(ctx context.Context, organizationID, id string, input UpdateInput) (Property, error) {
	property, err := scanProperty(r.pool.QueryRow(ctx, `
		UPDATE properties SET
			reference_code = COALESCE($3, reference_code),
			name = COALESCE($4, name),
			property_type = COALESCE($5, property_type),
			address_line_1 = COALESCE($6, address_line_1),
			address_line_2 = COALESCE($7, address_line_2),
			city = COALESCE($8, city),
			region = COALESCE($9, region),
			postal_code = COALESCE($10, postal_code),
			country_code = COALESCE($11, country_code),
			status = COALESCE($12, status),
			updated_at = now()
		WHERE organization_id = $1 AND id = $2
		RETURNING id, organization_id, reference_code, name, property_type,
		          address_line_1, address_line_2, city, region, postal_code,
		          country_code, status, external_avara_property_id, created_at, updated_at`,
		organizationID, id, input.ReferenceCode, input.Name, input.PropertyType,
		input.AddressLine1, input.AddressLine2, input.City, input.Region, input.PostalCode,
		input.CountryCode, input.Status,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Property{}, ErrNotFound
	}
	if err != nil {
		return Property{}, fmt.Errorf("update property: %w", err)
	}
	return property, nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanProperty(row rowScanner) (Property, error) {
	var property Property
	err := row.Scan(
		&property.ID, &property.OrganizationID, &property.ReferenceCode, &property.Name,
		&property.PropertyType, &property.AddressLine1, &property.AddressLine2, &property.City,
		&property.Region, &property.PostalCode, &property.CountryCode, &property.Status,
		&property.ExternalAvaraPropertyID, &property.CreatedAt, &property.UpdatedAt,
	)
	return property, err
}
