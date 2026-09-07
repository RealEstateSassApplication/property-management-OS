package notifications

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool: pool} }

func (r *PostgresRepository) List(ctx context.Context, organizationID string) ([]Notification, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, organization_id, COALESCE(actor_user_id::text,''), topic, channel, recipient, COALESCE(subject,''), body, payload, COALESCE(resource_type,''), COALESCE(resource_id::text,''), COALESCE(idempotency_key,''), status, attempt_count, max_attempts, available_at, locked_at, COALESCE(locked_by,''), COALESCE(last_error,''), delivered_at, created_at, updated_at FROM notification_outbox WHERE organization_id=$1 ORDER BY created_at DESC LIMIT 250`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Notification, 0)
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Enqueue(ctx context.Context, organizationID, actorUserID string, input EnqueueInput) (Notification, error) {
	row := r.pool.QueryRow(ctx, `INSERT INTO notification_outbox (organization_id, actor_user_id, topic, channel, recipient, subject, body, payload, resource_type, resource_id, idempotency_key) VALUES ($1, NULLIF($2,'')::uuid, $3, $4, $5, NULLIF($6,''), $7, COALESCE($8,'{}'::jsonb), NULLIF($9,''), NULLIF($10,'')::uuid, NULLIF($11,'')) ON CONFLICT DO NOTHING RETURNING id, organization_id, COALESCE(actor_user_id::text,''), topic, channel, recipient, COALESCE(subject,''), body, payload, COALESCE(resource_type,''), COALESCE(resource_id::text,''), COALESCE(idempotency_key,''), status, attempt_count, max_attempts, available_at, locked_at, COALESCE(locked_by,''), COALESCE(last_error,''), delivered_at, created_at, updated_at`, organizationID, actorUserID, input.Topic, input.Channel, input.Recipient, input.Subject, input.Body, input.Payload, input.ResourceType, input.ResourceID, input.IdempotencyKey)
	item, err := scanNotification(row)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) || input.IdempotencyKey == "" {
		return Notification{}, err
	}
	return scanNotification(r.pool.QueryRow(ctx, `SELECT id, organization_id, COALESCE(actor_user_id::text,''), topic, channel, recipient, COALESCE(subject,''), body, payload, COALESCE(resource_type,''), COALESCE(resource_id::text,''), COALESCE(idempotency_key,''), status, attempt_count, max_attempts, available_at, locked_at, COALESCE(locked_by,''), COALESCE(last_error,''), delivered_at, created_at, updated_at FROM notification_outbox WHERE organization_id=$1 AND idempotency_key=$2`, organizationID, input.IdempotencyKey))
}

func (r *PostgresRepository) GetRentReminderContext(ctx context.Context, organizationID, obligationID string) (RentReminderContext, error) {
	var item RentReminderContext
	err := r.pool.QueryRow(ctx, `SELECT ro.id, t.legal_name, p.name, u.label, ro.period, ro.due_date::text, ro.amount_minor-COALESCE(SUM(pa.amount_minor),0), ro.currency, CASE WHEN ro.state='void' THEN 'void' WHEN ro.amount_minor-COALESCE(SUM(pa.amount_minor),0)<=0 THEN 'paid' WHEN ro.due_date<CURRENT_DATE THEN 'overdue' ELSE 'open' END FROM rent_obligations ro JOIN leases l ON l.id=ro.lease_id AND l.organization_id=ro.organization_id JOIN tenancies tn ON tn.id=l.tenancy_id AND tn.organization_id=l.organization_id JOIN tenants t ON t.id=tn.primary_tenant_id AND t.organization_id=tn.organization_id JOIN units u ON u.id=tn.unit_id AND u.organization_id=tn.organization_id JOIN properties p ON p.id=u.property_id AND p.organization_id=u.organization_id LEFT JOIN payment_allocations pa ON pa.obligation_id=ro.id AND pa.organization_id=ro.organization_id WHERE ro.organization_id=$1 AND ro.id=$2 GROUP BY ro.id,t.legal_name,p.name,u.label,ro.period,ro.due_date,ro.amount_minor,ro.currency,ro.state`, organizationID, obligationID).Scan(&item.ObligationID, &item.TenantName, &item.PropertyName, &item.UnitLabel, &item.Period, &item.DueDate, &item.BalanceMinor, &item.Currency, &item.State)
	if errors.Is(err, pgx.ErrNoRows) {
		return RentReminderContext{}, ErrObligationNotFound
	}
	return item, err
}

func (r *PostgresRepository) ClaimBatch(ctx context.Context, workerID string, limit int) ([]Notification, error) {
	rows, err := r.pool.Query(ctx, `WITH candidates AS (SELECT id FROM notification_outbox WHERE (((status IN ('pending','retry')) AND available_at<=now()) OR (status='processing' AND locked_at<now()-interval '10 minutes')) AND attempt_count<max_attempts ORDER BY available_at, created_at FOR UPDATE SKIP LOCKED LIMIT $2) UPDATE notification_outbox n SET status='processing', locked_at=now(), locked_by=$1, attempt_count=n.attempt_count+1, updated_at=now() FROM candidates c WHERE n.id=c.id RETURNING n.id, n.organization_id, COALESCE(n.actor_user_id::text,''), n.topic, n.channel, n.recipient, COALESCE(n.subject,''), n.body, n.payload, COALESCE(n.resource_type,''), COALESCE(n.resource_id::text,''), COALESCE(n.idempotency_key,''), n.status, n.attempt_count, n.max_attempts, n.available_at, n.locked_at, COALESCE(n.locked_by,''), COALESCE(n.last_error,''), n.delivered_at, n.created_at, n.updated_at`, workerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Notification, 0)
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) MarkDelivered(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE notification_outbox SET status='delivered', delivered_at=now(), locked_at=NULL, locked_by=NULL, last_error=NULL, updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *PostgresRepository) MarkFailed(ctx context.Context, id, failure string, retryAt time.Time, dead bool) error {
	status := "retry"
	if dead {
		status = "dead"
	}
	_, err := r.pool.Exec(ctx, `UPDATE notification_outbox SET status=$2, available_at=$3, locked_at=NULL, locked_by=NULL, last_error=$4, updated_at=now() WHERE id=$1`, id, status, retryAt, failure)
	return err
}

type rowScanner interface{ Scan(dest ...any) error }

func scanNotification(row rowScanner) (Notification, error) {
	var item Notification
	err := row.Scan(&item.ID, &item.OrganizationID, &item.ActorUserID, &item.Topic, &item.Channel, &item.Recipient, &item.Subject, &item.Body, &item.Payload, &item.ResourceType, &item.ResourceID, &item.IdempotencyKey, &item.Status, &item.AttemptCount, &item.MaxAttempts, &item.AvailableAt, &item.LockedAt, &item.LockedBy, &item.LastError, &item.DeliveredAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
