package auth

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

func (r *PostgresRepository) GetMembership(ctx context.Context, organizationID, userID string) (Membership, error) {
	var membership Membership
	err := r.pool.QueryRow(ctx, `
		SELECT organization_id, user_id, role
		FROM organization_memberships
		WHERE organization_id = $1 AND user_id = $2
	`, organizationID, userID).Scan(&membership.OrganizationID, &membership.UserID, &membership.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, ErrMembershipNotFound
	}
	if err != nil {
		return Membership{}, err
	}
	return membership, nil
}

func (r *PostgresRepository) ResolveIdentity(ctx context.Context, issuer, subject, email string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		UPDATE user_identities
		SET last_seen_at = now(),
			last_seen_email = CASE WHEN NULLIF($3, '') IS NULL THEN last_seen_email ELSE $3 END
		WHERE issuer = $1 AND subject = $2
		RETURNING user_id
	`, issuer, subject, email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrIdentityNotLinked
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}
