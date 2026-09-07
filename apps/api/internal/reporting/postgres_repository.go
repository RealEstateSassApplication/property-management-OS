package reporting

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Dashboard(ctx context.Context, organizationID string) (Dashboard, error) {
	var item Dashboard
	if err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM properties WHERE organization_id=$1 AND status<>'archived'),
			(SELECT count(*) FROM units WHERE organization_id=$1),
			(SELECT count(*) FROM units WHERE organization_id=$1 AND occupancy_status='occupied'),
			(SELECT count(*) FROM maintenance_requests WHERE organization_id=$1 AND status NOT IN ('resolved','cancelled')),
			(SELECT count(*) FROM maintenance_requests WHERE organization_id=$1 AND status NOT IN ('resolved','cancelled') AND priority='emergency'),
			(SELECT count(*) FROM leases WHERE organization_id=$1 AND status='active' AND end_date>=CURRENT_DATE AND end_date<=CURRENT_DATE+30),
			(SELECT count(*) FROM leases WHERE organization_id=$1 AND status='active' AND end_date>=CURRENT_DATE AND end_date<=CURRENT_DATE+90)`, organizationID).Scan(
		&item.PropertyCount, &item.UnitCount, &item.OccupiedUnits, &item.OpenMaintenance,
		&item.EmergencyMaintenance, &item.LeasesExpiring30Days, &item.LeasesExpiring90Days,
	); err != nil {
		return Dashboard{}, err
	}
	if item.UnitCount > 0 {
		vacant := item.UnitCount - item.OccupiedUnits
		item.VacancyRateBPS = int((int64(vacant)*10000 + int64(item.UnitCount)/2) / int64(item.UnitCount))
	}
	var err error
	item.OutstandingByCurrency, err = r.obligationAmounts(ctx, organizationID, false)
	if err != nil {
		return Dashboard{}, err
	}
	item.OverdueByCurrency, err = r.obligationAmounts(ctx, organizationID, true)
	if err != nil {
		return Dashboard{}, err
	}
	item.CollectedMonthByCurrency, err = r.collectedThisMonth(ctx, organizationID)
	if err != nil {
		return Dashboard{}, err
	}
	return item, nil
}

func (r *PostgresRepository) ListAudit(ctx context.Context, organizationID string, limit int) ([]AuditEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, COALESCE(a.actor_user_id::text,''), COALESCE(u.display_name,''), a.action,
		       a.resource_type, COALESCE(a.resource_id::text,''), COALESCE(a.request_id,''), a.metadata, a.occurred_at
		FROM audit_events a
		LEFT JOIN users u ON u.id=a.actor_user_id
		WHERE a.organization_id=$1
		ORDER BY a.occurred_at DESC
		LIMIT $2`, organizationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AuditEvent, 0)
	for rows.Next() {
		var item AuditEvent
		if err := rows.Scan(&item.ID, &item.ActorUserID, &item.ActorName, &item.Action, &item.ResourceType, &item.ResourceID, &item.RequestID, &item.Metadata, &item.OccurredAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) obligationAmounts(ctx context.Context, organizationID string, overdueOnly bool) ([]CurrencyAmount, error) {
	condition := ""
	if overdueOnly {
		condition = " AND ro.due_date<CURRENT_DATE"
	}
	rows, err := r.pool.Query(ctx, `
		SELECT ro.currency, COALESCE(SUM(GREATEST(ro.amount_minor-COALESCE(a.allocated_minor,0),0)),0)::bigint
		FROM rent_obligations ro
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(pa.amount_minor),0)::bigint AS allocated_minor
			FROM payment_allocations pa
			WHERE pa.organization_id=ro.organization_id AND pa.obligation_id=ro.id
		) a ON true
		WHERE ro.organization_id=$1 AND ro.status<>'void'`+condition+`
		GROUP BY ro.currency
		HAVING SUM(GREATEST(ro.amount_minor-COALESCE(a.allocated_minor,0),0))>0
		ORDER BY ro.currency`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]CurrencyAmount, 0)
	for rows.Next() {
		var item CurrencyAmount
		if err := rows.Scan(&item.Currency, &item.AmountMinor); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) collectedThisMonth(ctx context.Context, organizationID string) ([]CurrencyAmount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT currency, COALESCE(SUM(amount_minor),0)::bigint
		FROM payments
		WHERE organization_id=$1 AND status='posted'
		  AND received_at>=date_trunc('month',CURRENT_DATE)::date
		  AND received_at<(date_trunc('month',CURRENT_DATE)+interval '1 month')::date
		GROUP BY currency ORDER BY currency`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]CurrencyAmount, 0)
	for rows.Next() {
		var item CurrencyAmount
		if err := rows.Scan(&item.Currency, &item.AmountMinor); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
