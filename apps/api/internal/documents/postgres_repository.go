package documents

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool: pool} }

const documentColumns = `id, organization_id, resource_type, resource_id, kind, file_name, content_type, size_bytes, storage_key, status, COALESCE(checksum_sha256, ''), COALESCE(uploaded_by_user_id::text, ''), verified_at, deleted_at, created_at, updated_at`

func scanDocument(row pgx.Row) (Document, error) {
	var item Document
	err := row.Scan(&item.ID, &item.OrganizationID, &item.ResourceType, &item.ResourceID, &item.Kind, &item.FileName, &item.ContentType, &item.SizeBytes, &item.StorageKey, &item.Status, &item.ChecksumSHA256, &item.UploadedByUserID, &item.VerifiedAt, &item.DeletedAt, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) CreatePending(ctx context.Context, item Document) (Document, error) {
	return scanDocument(r.pool.QueryRow(ctx, `
		INSERT INTO documents (id, organization_id, resource_type, resource_id, kind, file_name, content_type, size_bytes, storage_key, status, uploaded_by_user_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending',$10)
		RETURNING `+documentColumns,
		item.ID, item.OrganizationID, item.ResourceType, item.ResourceID, item.Kind, item.FileName, item.ContentType, item.SizeBytes, item.StorageKey, item.UploadedByUserID))
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, documentID string) (Document, error) {
	return scanDocument(r.pool.QueryRow(ctx, `SELECT `+documentColumns+` FROM documents WHERE organization_id=$1 AND id=$2`, organizationID, documentID))
}

func (r *PostgresRepository) List(ctx context.Context, organizationID string, filter Filter) ([]Document, error) {
	query := `SELECT ` + documentColumns + ` FROM documents WHERE organization_id=$1 AND status <> 'deleted'`
	args := []any{organizationID}
	if filter.ResourceType != "" {
		args = append(args, filter.ResourceType)
		query += fmt.Sprintf(" AND resource_type=$%d", len(args))
	}
	if filter.ResourceID != "" {
		args = append(args, filter.ResourceID)
		query += fmt.Sprintf(" AND resource_id=$%d", len(args))
	}
	query += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Document, 0)
	for rows.Next() {
		item, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ResourceExists(ctx context.Context, organizationID, resourceType, resourceID string) (bool, error) {
	var table string
	switch resourceType {
	case "property": table = "properties"
	case "unit": table = "units"
	case "tenant": table = "tenants"
	case "lease": table = "leases"
	case "owner": table = "owners"
	case "rent_payment": table = "payments"
	case "maintenance_request": table = "maintenance_requests"
	case "work_order": table = "work_orders"
	case "vendor": table = "vendors"
	default:
		return false, ErrInvalidInput
	}
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM `+table+` WHERE organization_id=$1 AND id=$2)`, organizationID, resourceID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, organizationID, documentID, actorUserID, status string, verifiedAt, deletedAt *time.Time) (Document, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil { return Document{}, err }
	defer tx.Rollback(ctx)
	item, err := scanDocument(tx.QueryRow(ctx, `
		UPDATE documents
		SET status=$3, verified_at=COALESCE($4, verified_at), deleted_at=COALESCE($5, deleted_at), updated_at=now()
		WHERE organization_id=$1 AND id=$2
		RETURNING `+documentColumns, organizationID, documentID, status, verifiedAt, deletedAt))
	if err != nil { return Document{}, err }
	_, err = tx.Exec(ctx, `INSERT INTO audit_events (organization_id, actor_user_id, action, resource_type, resource_id, metadata) VALUES ($1, NULLIF($2,'')::uuid, $3, 'document', $4, jsonb_build_object('status',$5))`, organizationID, actorUserID, "document.status_changed", documentID, status)
	if err != nil { return Document{}, err }
	if err := tx.Commit(ctx); err != nil { return Document{}, err }
	return item, nil
}
