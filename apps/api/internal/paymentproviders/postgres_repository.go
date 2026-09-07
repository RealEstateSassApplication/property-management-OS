package paymentproviders

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ProcessPaidEvent(ctx context.Context, provider, payloadHash string, event PaidEvent) (ProcessResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return ProcessResult{}, err
	}
	defer tx.Rollback(ctx)

	var eventRecordID string
	var createdAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO payment_provider_events (
			organization_id, provider, event_id, event_type, payload_sha256, status
		) VALUES ($1,$2,$3,$4,$5,'received')
		ON CONFLICT (provider,event_id) DO NOTHING
		RETURNING id,received_at
	`, event.OrganizationID, provider, event.EventID, event.EventType, payloadHash).Scan(&eventRecordID, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		var result ProcessResult
		result.Provider = provider
		result.EventID = event.EventID
		result.Duplicate = true
		err := tx.QueryRow(ctx, `
			SELECT status,COALESCE(payment_id::text,''),received_at
			FROM payment_provider_events
			WHERE provider=$1 AND event_id=$2
		`, provider, event.EventID).Scan(&result.Status, &result.PaymentID, &result.CreatedAt)
		if err != nil {
			return ProcessResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return ProcessResult{}, err
		}
		return result, nil
	}
	if err != nil {
		return ProcessResult{}, err
	}

	var tenantExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenants WHERE organization_id=$1 AND id=$2)`, event.OrganizationID, event.TenantID).Scan(&tenantExists); err != nil {
		return ProcessResult{}, err
	}
	if !tenantExists {
		if _, err := tx.Exec(ctx, `UPDATE payment_provider_events SET status='rejected',error_message=$3,processed_at=now() WHERE provider=$1 AND event_id=$2`, provider, event.EventID, ErrTenantNotFound.Error()); err != nil {
			return ProcessResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return ProcessResult{}, err
		}
		return ProcessResult{}, ErrTenantNotFound
	}

	var paymentID string
	err = tx.QueryRow(ctx, `
		INSERT INTO payments (
			organization_id,tenant_id,amount_minor,currency,received_at,method,reference_code,status
		) VALUES ($1,$2,$3,$4,$5::date,'online',$6,'posted')
		RETURNING id
	`, event.OrganizationID, event.TenantID, event.AmountMinor, event.Currency, event.ReceivedAt, event.ReferenceCode).Scan(&paymentID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if _, updateErr := tx.Exec(ctx, `UPDATE payment_provider_events SET status='rejected',error_message=$3,processed_at=now() WHERE provider=$1 AND event_id=$2`, provider, event.EventID, ErrDuplicateRef.Error()); updateErr != nil {
				return ProcessResult{}, updateErr
			}
			if commitErr := tx.Commit(ctx); commitErr != nil {
				return ProcessResult{}, commitErr
			}
			return ProcessResult{}, ErrDuplicateRef
		}
		return ProcessResult{}, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE payment_provider_events
		SET status='processed',payment_id=$3,processed_at=now(),error_message=NULL
		WHERE provider=$1 AND event_id=$2
	`, provider, event.EventID, paymentID); err != nil {
		return ProcessResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_events (organization_id,action,resource_type,resource_id,metadata)
		VALUES ($1,'payment.provider_received','payment',$2,jsonb_build_object('provider',$3,'eventId',$4,'referenceCode',$5))
	`, event.OrganizationID, paymentID, provider, event.EventID, event.ReferenceCode); err != nil {
		return ProcessResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ProcessResult{}, err
	}

	return ProcessResult{Provider: provider, EventID: event.EventID, Status: "processed", PaymentID: paymentID, CreatedAt: createdAt}, nil
}
