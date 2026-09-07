package accounting

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExpense(row scanner, item *PropertyExpense) error {
	return row.Scan(
		&item.ID, &item.OrganizationID, &item.PropertyID, &item.PropertyName, &item.VendorID, &item.VendorName,
		&item.WorkOrderID, &item.Category, &item.AmountMinor, &item.Currency, &item.IncurredOn, &item.Note,
		&item.ReferenceCode, &item.CreatedByUserID, &item.CreatedAt, &item.Reversed, &item.ReversalReason, &item.ReversedAt,
	)
}

func insertAudit(ctx context.Context, tx pgx.Tx, organizationID, actorUserID, action, resourceType, resourceID, metadata string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_events (organization_id,actor_user_id,action,resource_type,resource_id,metadata)
		VALUES ($1,NULLIF($2,'')::uuid,$3,$4,NULLIF($5,'')::uuid,$6::jsonb)
	`, organizationID, actorUserID, action, resourceType, resourceID, metadata)
	return err
}
