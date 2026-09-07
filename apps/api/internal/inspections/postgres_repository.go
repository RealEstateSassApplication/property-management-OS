package inspections

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const inspectionSelect = `
	SELECT i.id,i.organization_id,i.tenancy_id,i.property_id,p.name,i.unit_id,u.label,
	       COALESCE(t.legal_name,''),i.inspection_type,i.status,
	       COALESCE(to_char(i.scheduled_for AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
	       COALESCE(i.summary,''),
	       COALESCE(to_char(i.completed_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
	       COALESCE(to_char(i.acknowledged_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
	       COALESCE(i.completed_by_user_id::text,''),COALESCE(i.acknowledged_by_user_id::text,''),
	       i.created_at,i.updated_at
	FROM inspections i
	JOIN properties p ON p.id=i.property_id AND p.organization_id=i.organization_id
	JOIN units u ON u.id=i.unit_id AND u.organization_id=i.organization_id
	LEFT JOIN tenancy_tenants tt ON tt.tenancy_id=i.tenancy_id AND tt.organization_id=i.organization_id AND tt.role='primary'
	LEFT JOIN tenants t ON t.id=tt.tenant_id AND t.organization_id=i.organization_id
`

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]Inspection, error) {
	rows, err := r.pool.Query(ctx, inspectionSelect+` WHERE i.organization_id=$1 ORDER BY COALESCE(i.scheduled_for,i.created_at) DESC`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Inspection, 0)
	for rows.Next() {
		var item Inspection
		if err := scanInspection(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, organizationID, inspectionID string) (Inspection, error) {
	var item Inspection
	err := scanInspection(r.pool.QueryRow(ctx, inspectionSelect+` WHERE i.organization_id=$1 AND i.id=$2`, organizationID, inspectionID), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return Inspection{}, ErrInspectionNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListItems(ctx context.Context, organizationID, inspectionID string) ([]Item, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,organization_id,inspection_id,area,item_name,condition,COALESCE(notes,''),COALESCE(evidence_document_id::text,''),created_at,updated_at FROM inspection_items WHERE organization_id=$1 AND inspection_id=$2 ORDER BY area,item_name,created_at`, organizationID, inspectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.InspectionID, &item.Area, &item.ItemName, &item.Condition, &item.Notes, &item.EvidenceDocumentID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, organizationID, actorUserID string, input CreateInspectionInput) (Inspection, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Inspection{}, err
	}
	defer tx.Rollback(ctx)

	var propertyID, unitID string
	err = tx.QueryRow(ctx, `SELECT u.property_id,tn.unit_id FROM tenancies tn JOIN units u ON u.id=tn.unit_id AND u.organization_id=tn.organization_id WHERE tn.organization_id=$1 AND tn.id=$2`, organizationID, input.TenancyID).Scan(&propertyID, &unitID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Inspection{}, ErrTenancyNotFound
	}
	if err != nil {
		return Inspection{}, err
	}

	var scheduled any
	if input.ScheduledFor != "" {
		parsed, err := time.Parse(time.RFC3339, input.ScheduledFor)
		if err != nil {
			return Inspection{}, err
		}
		scheduled = parsed
	}

	var inspectionID string
	if err := tx.QueryRow(ctx, `INSERT INTO inspections (organization_id,tenancy_id,property_id,unit_id,inspection_type,status,scheduled_for,summary,created_by_user_id) VALUES ($1,$2,$3,$4,$5,'draft',$6,NULLIF($7,''),NULLIF($8,'')::uuid) RETURNING id`, organizationID, input.TenancyID, propertyID, unitID, input.InspectionType, scheduled, input.Summary, actorUserID).Scan(&inspectionID); err != nil {
		return Inspection{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_events (organization_id,actor_user_id,action,resource_type,resource_id,metadata) VALUES ($1,NULLIF($2,'')::uuid,'inspection.created','inspection',$3,jsonb_build_object('type',$4,'tenancyId',$5))`, organizationID, actorUserID, inspectionID, input.InspectionType, input.TenancyID); err != nil {
		return Inspection{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Inspection{}, err
	}
	return r.Get(ctx, organizationID, inspectionID)
}

func (r *PostgresRepository) CreateItem(ctx context.Context, organizationID string, input CreateItemInput) (Item, error) {
	var status string
	if err := r.pool.QueryRow(ctx, `SELECT status FROM inspections WHERE organization_id=$1 AND id=$2`, organizationID, input.InspectionID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Item{}, ErrInspectionNotFound
		}
		return Item{}, err
	}
	if status == "completed" || status == "acknowledged" || status == "cancelled" {
		return Item{}, ErrInspectionClosed
	}

	if input.EvidenceDocumentID != "" {
		var docStatus string
		if err := r.pool.QueryRow(ctx, `SELECT status FROM documents WHERE organization_id=$1 AND id=$2`, organizationID, input.EvidenceDocumentID).Scan(&docStatus); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Item{}, ErrDocumentNotFound
			}
			return Item{}, err
		}
		if docStatus != "available" {
			return Item{}, ErrDocumentNotAvailable
		}
	}

	var item Item
	err := r.pool.QueryRow(ctx, `INSERT INTO inspection_items (organization_id,inspection_id,area,item_name,condition,notes,evidence_document_id) VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,'')::uuid) RETURNING id,organization_id,inspection_id,area,item_name,condition,COALESCE(notes,''),COALESCE(evidence_document_id::text,''),created_at,updated_at`, organizationID, input.InspectionID, input.Area, input.ItemName, input.Condition, input.Notes, input.EvidenceDocumentID).Scan(&item.ID, &item.OrganizationID, &item.InspectionID, &item.Area, &item.ItemName, &item.Condition, &item.Notes, &item.EvidenceDocumentID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Item{}, err
	}
	_, _ = r.pool.Exec(ctx, `UPDATE inspections SET status='in_progress',updated_at=now() WHERE organization_id=$1 AND id=$2 AND status='draft'`, organizationID, input.InspectionID)
	return item, nil
}

func (r *PostgresRepository) Complete(ctx context.Context, organizationID, actorUserID string, input CompleteInspectionInput) (Inspection, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Inspection{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM inspections WHERE organization_id=$1 AND id=$2 FOR UPDATE`, organizationID, input.InspectionID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Inspection{}, ErrInspectionNotFound
		}
		return Inspection{}, err
	}
	if status == "completed" || status == "acknowledged" || status == "cancelled" {
		return Inspection{}, ErrInspectionClosed
	}

	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM inspection_items WHERE organization_id=$1 AND inspection_id=$2`, organizationID, input.InspectionID).Scan(&count); err != nil {
		return Inspection{}, err
	}
	if count == 0 {
		return Inspection{}, ErrInspectionHasNoItems
	}

	if _, err := tx.Exec(ctx, `UPDATE inspections SET status='completed',summary=COALESCE(NULLIF($3,''),summary),completed_at=now(),completed_by_user_id=NULLIF($4,'')::uuid,updated_at=now() WHERE organization_id=$1 AND id=$2`, organizationID, input.InspectionID, input.Summary, actorUserID); err != nil {
		return Inspection{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_events (organization_id,actor_user_id,action,resource_type,resource_id,metadata) VALUES ($1,NULLIF($2,'')::uuid,'inspection.completed','inspection',$3,jsonb_build_object('itemCount',$4))`, organizationID, actorUserID, input.InspectionID, count); err != nil {
		return Inspection{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Inspection{}, err
	}
	return r.Get(ctx, organizationID, input.InspectionID)
}

func (r *PostgresRepository) Acknowledge(ctx context.Context, organizationID, actorUserID string, input AcknowledgeInspectionInput) (Inspection, error) {
	result, err := r.pool.Exec(ctx, `UPDATE inspections SET status='acknowledged',acknowledged_at=now(),acknowledged_by_user_id=NULLIF($3,'')::uuid,updated_at=now() WHERE organization_id=$1 AND id=$2 AND status='completed'`, organizationID, input.InspectionID, actorUserID)
	if err != nil {
		return Inspection{}, err
	}
	if result.RowsAffected() == 0 {
		var status string
		if err := r.pool.QueryRow(ctx, `SELECT status FROM inspections WHERE organization_id=$1 AND id=$2`, organizationID, input.InspectionID).Scan(&status); errors.Is(err, pgx.ErrNoRows) {
			return Inspection{}, ErrInspectionNotFound
		} else if err != nil {
			return Inspection{}, err
		}
		return Inspection{}, ErrInspectionNotCompleted
	}
	_, _ = r.pool.Exec(ctx, `INSERT INTO audit_events (organization_id,actor_user_id,action,resource_type,resource_id,metadata) VALUES ($1,NULLIF($2,'')::uuid,'inspection.acknowledged','inspection',$3,'{}'::jsonb)`, organizationID, actorUserID, input.InspectionID)
	return r.Get(ctx, organizationID, input.InspectionID)
}

type scanner interface{ Scan(dest ...any) error }

func scanInspection(row scanner, item *Inspection) error {
	return row.Scan(&item.ID, &item.OrganizationID, &item.TenancyID, &item.PropertyID, &item.PropertyName, &item.UnitID, &item.UnitLabel, &item.PrimaryTenantName, &item.InspectionType, &item.Status, &item.ScheduledFor, &item.Summary, &item.CompletedAt, &item.AcknowledgedAt, &item.CompletedByUserID, &item.AcknowledgedByUserID, &item.CreatedAt, &item.UpdatedAt)
}
