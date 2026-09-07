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
