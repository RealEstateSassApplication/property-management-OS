package tenants

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

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]Tenant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, organization_id, legal_name, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
		FROM tenants
		WHERE organization_id = $1
		ORDER BY legal_name ASC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Tenant, 0)
	for rows.Next() {
		var tenant Tenant
		if err := rows.Scan(&tenant.ID, &tenant.OrganizationID, &tenant.LegalName, &tenant.Email, &tenant.Phone, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, tenant)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, tenantID string) (Tenant, error) {
	var tenant Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, organization_id, legal_name, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
		FROM tenants
		WHERE organization_id = $1 AND id = $2
	`, organizationID, tenantID).Scan(&tenant.ID, &tenant.OrganizationID, &tenant.LegalName, &tenant.Email, &tenant.Phone, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrNotFound
	}
	return tenant, err
}

func (r *PostgresRepository) Create(ctx context.Context, organizationID string, input CreateInput) (Tenant, error) {
	var tenant Tenant
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tenants (organization_id, legal_name, email, phone, status)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5)
		RETURNING id, organization_id, legal_name, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
	`, organizationID, input.LegalName, input.Email, input.Phone, input.Status).Scan(
		&tenant.ID, &tenant.OrganizationID, &tenant.LegalName, &tenant.Email, &tenant.Phone, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	return tenant, err
}

func (r *PostgresRepository) Update(ctx context.Context, organizationID, tenantID string, input UpdateInput) (Tenant, error) {
	var tenant Tenant
	legalNameSet := input.LegalName != nil
	emailSet := input.Email != nil
	phoneSet := input.Phone != nil
	statusSet := input.Status != nil
	legalName, email, phone, status := "", "", "", ""
	if input.LegalName != nil {
		legalName = *input.LegalName
	}
	if input.Email != nil {
		email = *input.Email
	}
	if input.Phone != nil {
		phone = *input.Phone
	}
	if input.Status != nil {
		status = *input.Status
	}

	err := r.pool.QueryRow(ctx, `
		UPDATE tenants
		SET legal_name = CASE WHEN $3 THEN $4 ELSE legal_name END,
			email = CASE WHEN $5 THEN NULLIF($6, '') ELSE email END,
			phone = CASE WHEN $7 THEN NULLIF($8, '') ELSE phone END,
			status = CASE WHEN $9 THEN $10 ELSE status END,
			updated_at = now()
		WHERE organization_id = $1 AND id = $2
		RETURNING id, organization_id, legal_name, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
	`, organizationID, tenantID, legalNameSet, legalName, emailSet, email, phoneSet, phone, statusSet, status).Scan(
		&tenant.ID, &tenant.OrganizationID, &tenant.LegalName, &tenant.Email, &tenant.Phone, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrNotFound
	}
	return tenant, err
}
