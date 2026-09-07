package owners

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

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]Owner, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, organization_id, legal_name, owner_type, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
		FROM owners
		WHERE organization_id = $1
		ORDER BY legal_name, created_at
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Owner, 0)
	for rows.Next() {
		var owner Owner
		if err := rows.Scan(&owner.ID, &owner.OrganizationID, &owner.LegalName, &owner.OwnerType, &owner.Email, &owner.Phone, &owner.Status, &owner.CreatedAt, &owner.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, owner)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, ownerID string) (Owner, error) {
	var owner Owner
	err := r.pool.QueryRow(ctx, `
		SELECT id, organization_id, legal_name, owner_type, COALESCE(email, ''), COALESCE(phone, ''), status, created_at, updated_at
		FROM owners WHERE organization_id = $1 AND id = $2
	`, organizationID, ownerID).Scan(&owner.ID, &owner.OrganizationID, &owner.LegalName, &owner.OwnerType, &owner.Email, &owner.Phone, &owner.Status, &owner.CreatedAt, &owner.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Owner{}, ErrNotFound
	}
	return owner, err
}

func (r *PostgresRepository) Create(ctx context.Context, organizationID string, input CreateOwnerInput) (Owner, error) {
	var ownerID string
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO owners (organization_id, legal_name, owner_type, email, phone, status)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6)
		RETURNING id
	`, organizationID, input.LegalName, input.OwnerType, input.Email, input.Phone, input.Status).Scan(&ownerID); err != nil {
		return Owner{}, err
	}
	return r.Get(ctx, organizationID, ownerID)
}

func (r *PostgresRepository) ListInterests(ctx context.Context, organizationID string) ([]OwnershipInterest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT oi.id, oi.organization_id, oi.property_id, p.name, oi.owner_id, o.legal_name,
			oi.ownership_bps, to_char(oi.effective_from, 'YYYY-MM-DD'), COALESCE(to_char(oi.effective_to, 'YYYY-MM-DD'), ''), oi.created_at
		FROM ownership_interests oi
		JOIN properties p ON p.id = oi.property_id AND p.organization_id = oi.organization_id
		JOIN owners o ON o.id = oi.owner_id AND o.organization_id = oi.organization_id
		WHERE oi.organization_id = $1
		ORDER BY p.name, oi.effective_from DESC, o.legal_name
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]OwnershipInterest, 0)
	for rows.Next() {
		var interest OwnershipInterest
		if err := rows.Scan(&interest.ID, &interest.OrganizationID, &interest.PropertyID, &interest.PropertyName, &interest.OwnerID, &interest.OwnerName, &interest.OwnershipBps, &interest.EffectiveFrom, &interest.EffectiveTo, &interest.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, interest)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) CreateInterest(ctx context.Context, organizationID string, input CreateInterestInput) (OwnershipInterest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return OwnershipInterest{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var ownerExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM owners WHERE organization_id = $1 AND id = $2)`, organizationID, input.OwnerID).Scan(&ownerExists); err != nil {
		return OwnershipInterest{}, err
	}
	if !ownerExists {
		return OwnershipInterest{}, ErrOwnerNotFound
	}

	var lockedPropertyID string
	if err := tx.QueryRow(ctx, `
		SELECT id FROM properties
		WHERE organization_id = $1 AND id = $2
		FOR UPDATE
	`, organizationID, input.PropertyID).Scan(&lockedPropertyID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OwnershipInterest{}, ErrPropertyNotFound
		}
		return OwnershipInterest{}, err
	}

	var currentForOwner bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM ownership_interests
			WHERE organization_id = $1 AND property_id = $2 AND owner_id = $3 AND effective_to IS NULL
		)
	`, organizationID, input.PropertyID, input.OwnerID).Scan(&currentForOwner); err != nil {
		return OwnershipInterest{}, err
	}
	if currentForOwner {
		return OwnershipInterest{}, ErrCurrentInterestExists
	}

	var currentBps int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(ownership_bps), 0)
		FROM ownership_interests
		WHERE organization_id = $1 AND property_id = $2 AND effective_to IS NULL
	`, organizationID, input.PropertyID).Scan(&currentBps); err != nil {
		return OwnershipInterest{}, err
	}
	if currentBps+input.OwnershipBps > 10000 {
		return OwnershipInterest{}, ErrOwnershipExceeded
	}

	var interestID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO ownership_interests (organization_id, property_id, owner_id, ownership_bps, effective_from)
		VALUES ($1, $2, $3, $4, $5::date)
		RETURNING id
	`, organizationID, input.PropertyID, input.OwnerID, input.OwnershipBps, input.EffectiveFrom).Scan(&interestID); err != nil {
		return OwnershipInterest{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return OwnershipInterest{}, err
	}

	var interest OwnershipInterest
	err = r.pool.QueryRow(ctx, `
		SELECT oi.id, oi.organization_id, oi.property_id, p.name, oi.owner_id, o.legal_name,
			oi.ownership_bps, to_char(oi.effective_from, 'YYYY-MM-DD'), COALESCE(to_char(oi.effective_to, 'YYYY-MM-DD'), ''), oi.created_at
		FROM ownership_interests oi
		JOIN properties p ON p.id = oi.property_id AND p.organization_id = oi.organization_id
		JOIN owners o ON o.id = oi.owner_id AND o.organization_id = oi.organization_id
		WHERE oi.organization_id = $1 AND oi.id = $2
	`, organizationID, interestID).Scan(&interest.ID, &interest.OrganizationID, &interest.PropertyID, &interest.PropertyName, &interest.OwnerID, &interest.OwnerName, &interest.OwnershipBps, &interest.EffectiveFrom, &interest.EffectiveTo, &interest.CreatedAt)
	return interest, err
}
