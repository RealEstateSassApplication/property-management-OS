package orgadmin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetSettings(ctx context.Context, organizationID string) (OrganizationSettings, error) {
	var item OrganizationSettings
	err := r.pool.QueryRow(ctx, `
		SELECT o.id, o.name, o.slug, o.status,
		       COALESCE(s.timezone,'Asia/Colombo'), COALESCE(s.default_currency,'LKR'),
		       COALESCE(s.country_code,'LK'), COALESCE(s.billing_email,''),
		       COALESCE(s.updated_at,o.updated_at)
		FROM organizations o
		LEFT JOIN organization_settings s ON s.organization_id=o.id
		WHERE o.id=$1`, organizationID).Scan(
		&item.OrganizationID, &item.Name, &item.Slug, &item.Status,
		&item.Timezone, &item.DefaultCurrency, &item.CountryCode, &item.BillingEmail, &item.UpdatedAt,
	)
	return item, err
}

func (r *PostgresRepository) UpdateSettings(ctx context.Context, organizationID string, input UpdateOrganizationInput) (OrganizationSettings, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return OrganizationSettings{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE organizations SET name=$2, updated_at=now() WHERE id=$1`, organizationID, input.Name); err != nil {
		return OrganizationSettings{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_settings (organization_id, timezone, default_currency, country_code, billing_email)
		VALUES ($1,$2,$3,$4,NULLIF($5,''))
		ON CONFLICT (organization_id) DO UPDATE SET
			timezone=EXCLUDED.timezone,
			default_currency=EXCLUDED.default_currency,
			country_code=EXCLUDED.country_code,
			billing_email=EXCLUDED.billing_email,
			updated_at=now()`, organizationID, input.Timezone, input.DefaultCurrency, input.CountryCode, input.BillingEmail); err != nil {
		return OrganizationSettings{}, err
	}
	var item OrganizationSettings
	if err := tx.QueryRow(ctx, `
		SELECT o.id,o.name,o.slug,o.status,s.timezone,s.default_currency,s.country_code,COALESCE(s.billing_email,''),s.updated_at
		FROM organizations o JOIN organization_settings s ON s.organization_id=o.id WHERE o.id=$1`, organizationID).Scan(
		&item.OrganizationID, &item.Name, &item.Slug, &item.Status, &item.Timezone,
		&item.DefaultCurrency, &item.CountryCode, &item.BillingEmail, &item.UpdatedAt,
	); err != nil {
		return OrganizationSettings{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return OrganizationSettings{}, err
	}
	return item, nil
}

func (r *PostgresRepository) ListMembers(ctx context.Context, organizationID string) ([]Member, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id,u.email,u.display_name,u.status,m.role,m.created_at
		FROM organization_memberships m JOIN users u ON u.id=m.user_id
		WHERE m.organization_id=$1
		ORDER BY CASE m.role WHEN 'admin' THEN 0 WHEN 'manager' THEN 1 ELSE 2 END, lower(u.display_name), lower(u.email)`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Member, 0)
	for rows.Next() {
		var item Member
		if err := rows.Scan(&item.UserID, &item.Email, &item.DisplayName, &item.UserStatus, &item.Role, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetMember(ctx context.Context, organizationID, userID string) (Member, error) {
	var item Member
	err := r.pool.QueryRow(ctx, `
		SELECT u.id,u.email,u.display_name,u.status,m.role,m.created_at
		FROM organization_memberships m JOIN users u ON u.id=m.user_id
		WHERE m.organization_id=$1 AND m.user_id=$2`, organizationID, userID).Scan(
		&item.UserID, &item.Email, &item.DisplayName, &item.UserStatus, &item.Role, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}
	return item, err
}

func (r *PostgresRepository) UpsertMember(ctx context.Context, organizationID string, input CreateMemberInput) (Member, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Member{}, err
	}
	defer tx.Rollback(ctx)
	var userID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (email, display_name, status)
		VALUES ($1,$2,'invited')
		ON CONFLICT (email) DO UPDATE SET
			display_name=CASE WHEN users.status='invited' THEN EXCLUDED.display_name ELSE users.display_name END
		RETURNING id`, input.Email, input.DisplayName).Scan(&userID); err != nil {
		return Member{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_memberships (organization_id,user_id,role)
		VALUES ($1,$2,$3)
		ON CONFLICT (organization_id,user_id) DO UPDATE SET role=EXCLUDED.role`, organizationID, userID, input.Role); err != nil {
		return Member{}, err
	}
	if input.Role != "owner" {
		if _, err := tx.Exec(ctx, `DELETE FROM owner_user_links WHERE organization_id=$1 AND user_id=$2`, organizationID, userID); err != nil {
			return Member{}, err
		}
	}
	if input.Role != "tenant" {
		if _, err := tx.Exec(ctx, `DELETE FROM tenant_user_links WHERE organization_id=$1 AND user_id=$2`, organizationID, userID); err != nil {
			return Member{}, err
		}
	}
	var item Member
	if err := tx.QueryRow(ctx, `
		SELECT u.id,u.email,u.display_name,u.status,m.role,m.created_at
		FROM organization_memberships m JOIN users u ON u.id=m.user_id
		WHERE m.organization_id=$1 AND m.user_id=$2`, organizationID, userID).Scan(
		&item.UserID, &item.Email, &item.DisplayName, &item.UserStatus, &item.Role, &item.CreatedAt,
	); err != nil {
		return Member{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Member{}, err
	}
	return item, nil
}

func (r *PostgresRepository) UpdateMemberRole(ctx context.Context, organizationID, userID, role string) (Member, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Member{}, err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `UPDATE organization_memberships SET role=$3 WHERE organization_id=$1 AND user_id=$2`, organizationID, userID, role)
	if err != nil {
		return Member{}, err
	}
	if result.RowsAffected() == 0 {
		return Member{}, ErrMemberNotFound
	}
	if role != "owner" {
		if _, err := tx.Exec(ctx, `DELETE FROM owner_user_links WHERE organization_id=$1 AND user_id=$2`, organizationID, userID); err != nil {
			return Member{}, err
		}
	}
	if role != "tenant" {
		if _, err := tx.Exec(ctx, `DELETE FROM tenant_user_links WHERE organization_id=$1 AND user_id=$2`, organizationID, userID); err != nil {
			return Member{}, err
		}
	}
	var item Member
	if err := tx.QueryRow(ctx, `
		SELECT u.id,u.email,u.display_name,u.status,m.role,m.created_at
		FROM organization_memberships m JOIN users u ON u.id=m.user_id
		WHERE m.organization_id=$1 AND m.user_id=$2`, organizationID, userID).Scan(
		&item.UserID, &item.Email, &item.DisplayName, &item.UserStatus, &item.Role, &item.CreatedAt,
	); err != nil {
		return Member{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Member{}, err
	}
	return item, nil
}

func (r *PostgresRepository) RemoveMember(ctx context.Context, organizationID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM owner_user_links WHERE organization_id=$1 AND user_id=$2`, organizationID, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM tenant_user_links WHERE organization_id=$1 AND user_id=$2`, organizationID, userID); err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `DELETE FROM organization_memberships WHERE organization_id=$1 AND user_id=$2`, organizationID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) CountAdmins(ctx context.Context, organizationID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM organization_memberships WHERE organization_id=$1 AND role='admin'`, organizationID).Scan(&count)
	return count, err
}

func (r *PostgresRepository) ListPortalLinks(ctx context.Context, organizationID string) ([]PortalLink, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 'owner',u.id,u.email,u.display_name,o.id,o.legal_name
		FROM owner_user_links l JOIN users u ON u.id=l.user_id JOIN owners o ON o.id=l.owner_id AND o.organization_id=l.organization_id
		WHERE l.organization_id=$1
		UNION ALL
		SELECT 'tenant',u.id,u.email,u.display_name,t.id,t.legal_name
		FROM tenant_user_links l JOIN users u ON u.id=l.user_id JOIN tenants t ON t.id=l.tenant_id AND t.organization_id=l.organization_id
		WHERE l.organization_id=$1
		ORDER BY 1,4`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PortalLink, 0)
	for rows.Next() {
		var item PortalLink
		if err := rows.Scan(&item.Kind, &item.UserID, &item.Email, &item.DisplayName, &item.ResourceID, &item.ResourceName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) LinkOwner(ctx context.Context, organizationID string, input CreateOwnerPortalLinkInput) (PortalLink, error) {
	result, err := r.pool.Exec(ctx, `
		INSERT INTO owner_user_links (organization_id,owner_id,user_id)
		SELECT m.organization_id,o.id,m.user_id
		FROM organization_memberships m
		JOIN owners o ON o.organization_id=m.organization_id AND o.id=$3
		WHERE m.organization_id=$1 AND m.user_id=$2 AND m.role='owner'
		ON CONFLICT DO NOTHING`, organizationID, input.UserID, input.OwnerID)
	if err != nil {
		return PortalLink{}, err
	}
	if result.RowsAffected() == 0 {
		var exists bool
		_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM owner_user_links WHERE organization_id=$1 AND user_id=$2 AND owner_id=$3)`, organizationID, input.UserID, input.OwnerID).Scan(&exists)
		if !exists {
			return PortalLink{}, ErrPortalLinkInvalid
		}
	}
	return r.getOwnerLink(ctx, organizationID, input.UserID, input.OwnerID)
}

func (r *PostgresRepository) LinkTenant(ctx context.Context, organizationID string, input CreateTenantPortalLinkInput) (PortalLink, error) {
	result, err := r.pool.Exec(ctx, `
		INSERT INTO tenant_user_links (organization_id,tenant_id,user_id)
		SELECT m.organization_id,t.id,m.user_id
		FROM organization_memberships m
		JOIN tenants t ON t.organization_id=m.organization_id AND t.id=$3
		WHERE m.organization_id=$1 AND m.user_id=$2 AND m.role='tenant'
		ON CONFLICT DO NOTHING`, organizationID, input.UserID, input.TenantID)
	if err != nil {
		return PortalLink{}, err
	}
	if result.RowsAffected() == 0 {
		var exists bool
		_ = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_user_links WHERE organization_id=$1 AND user_id=$2 AND tenant_id=$3)`, organizationID, input.UserID, input.TenantID).Scan(&exists)
		if !exists {
			return PortalLink{}, ErrPortalLinkInvalid
		}
	}
	return r.getTenantLink(ctx, organizationID, input.UserID, input.TenantID)
}

func (r *PostgresRepository) UnlinkOwner(ctx context.Context, organizationID, userID, ownerID string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM owner_user_links WHERE organization_id=$1 AND user_id=$2 AND owner_id=$3`, organizationID, userID, ownerID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPortalLinkInvalid
	}
	return nil
}

func (r *PostgresRepository) UnlinkTenant(ctx context.Context, organizationID, userID, tenantID string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM tenant_user_links WHERE organization_id=$1 AND user_id=$2 AND tenant_id=$3`, organizationID, userID, tenantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPortalLinkInvalid
	}
	return nil
}

func (r *PostgresRepository) getOwnerLink(ctx context.Context, organizationID, userID, ownerID string) (PortalLink, error) {
	var item PortalLink
	item.Kind = "owner"
	err := r.pool.QueryRow(ctx, `
		SELECT u.id,u.email,u.display_name,o.id,o.legal_name
		FROM owner_user_links l JOIN users u ON u.id=l.user_id JOIN owners o ON o.id=l.owner_id AND o.organization_id=l.organization_id
		WHERE l.organization_id=$1 AND l.user_id=$2 AND l.owner_id=$3`, organizationID, userID, ownerID).Scan(
		&item.UserID, &item.Email, &item.DisplayName, &item.ResourceID, &item.ResourceName,
	)
	return item, err
}

func (r *PostgresRepository) getTenantLink(ctx context.Context, organizationID, userID, tenantID string) (PortalLink, error) {
	var item PortalLink
	item.Kind = "tenant"
	err := r.pool.QueryRow(ctx, `
		SELECT u.id,u.email,u.display_name,t.id,t.legal_name
		FROM tenant_user_links l JOIN users u ON u.id=l.user_id JOIN tenants t ON t.id=l.tenant_id AND t.organization_id=l.organization_id
		WHERE l.organization_id=$1 AND l.user_id=$2 AND l.tenant_id=$3`, organizationID, userID, tenantID).Scan(
		&item.UserID, &item.Email, &item.DisplayName, &item.ResourceID, &item.ResourceName,
	)
	return item, err
}
