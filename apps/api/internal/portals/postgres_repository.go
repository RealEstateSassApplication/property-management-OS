package portals

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

func (r *PostgresRepository) ListOwnerProfiles(ctx context.Context, organizationID, userID string) ([]OwnerProfile, error) {
	rows, err := r.pool.Query(ctx, `SELECT o.id, o.legal_name, o.owner_type, COALESCE(o.email,''), COALESCE(o.phone,'') FROM owner_user_links l JOIN owners o ON o.id=l.owner_id AND o.organization_id=l.organization_id WHERE l.organization_id=$1 AND l.user_id=$2 AND o.status='active' ORDER BY o.legal_name`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OwnerProfile, 0)
	for rows.Next() {
		var item OwnerProfile
		if err := rows.Scan(&item.ID, &item.LegalName, &item.OwnerType, &item.Email, &item.Phone); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListOwnerPropertyBases(ctx context.Context, organizationID, userID string) ([]OwnerPropertyBase, error) {
	rows, err := r.pool.Query(ctx, `SELECT oi.owner_id, p.id, p.reference_code, p.name, COALESCE(p.city,''), oi.ownership_bps, (SELECT count(*) FROM units u WHERE u.organization_id=p.organization_id AND u.property_id=p.id), (SELECT count(*) FROM units u WHERE u.organization_id=p.organization_id AND u.property_id=p.id AND u.occupancy_status='occupied'), (SELECT count(*) FROM maintenance_requests mr WHERE mr.organization_id=p.organization_id AND mr.property_id=p.id AND mr.status NOT IN ('resolved','cancelled')) FROM owner_user_links l JOIN ownership_interests oi ON oi.organization_id=l.organization_id AND oi.owner_id=l.owner_id JOIN properties p ON p.organization_id=oi.organization_id AND p.id=oi.property_id WHERE l.organization_id=$1 AND l.user_id=$2 AND oi.effective_from<=CURRENT_DATE AND (oi.effective_to IS NULL OR oi.effective_to>=CURRENT_DATE) AND p.status<>'archived' ORDER BY p.name, oi.owner_id`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OwnerPropertyBase, 0)
	for rows.Next() {
		var item OwnerPropertyBase
		if err := rows.Scan(&item.OwnerID, &item.PropertyID, &item.ReferenceCode, &item.Name, &item.City, &item.OwnershipBPS, &item.UnitCount, &item.OccupiedUnits, &item.OpenMaintenance); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListOwnerReceivables(ctx context.Context, organizationID, userID string) ([]OwnerReceivableRow, error) {
	rows, err := r.pool.Query(ctx, `WITH linked_properties AS (SELECT DISTINCT oi.property_id FROM owner_user_links l JOIN ownership_interests oi ON oi.organization_id=l.organization_id AND oi.owner_id=l.owner_id WHERE l.organization_id=$1 AND l.user_id=$2 AND oi.effective_from<=CURRENT_DATE AND (oi.effective_to IS NULL OR oi.effective_to>=CURRENT_DATE)) SELECT lp.property_id, ro.currency, COALESCE(SUM(GREATEST(ro.amount_minor-COALESCE(a.allocated_minor,0),0)),0)::bigint FROM linked_properties lp JOIN units u ON u.organization_id=$1 AND u.property_id=lp.property_id JOIN tenancies tn ON tn.organization_id=u.organization_id AND tn.unit_id=u.id JOIN leases l ON l.organization_id=tn.organization_id AND l.tenancy_id=tn.id JOIN rent_obligations ro ON ro.organization_id=l.organization_id AND ro.lease_id=l.id LEFT JOIN LATERAL (SELECT COALESCE(SUM(pa.amount_minor),0)::bigint AS allocated_minor FROM payment_allocations pa WHERE pa.organization_id=ro.organization_id AND pa.obligation_id=ro.id) a ON true WHERE ro.status<>'void' GROUP BY lp.property_id, ro.currency ORDER BY lp.property_id, ro.currency`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OwnerReceivableRow, 0)
	for rows.Next() {
		var item OwnerReceivableRow
		if err := rows.Scan(&item.PropertyID, &item.Currency, &item.OutstandingMinor); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListTenantProfiles(ctx context.Context, organizationID, userID string) ([]TenantProfile, error) {
	rows, err := r.pool.Query(ctx, `SELECT t.id, t.legal_name, COALESCE(t.email,''), COALESCE(t.phone,''), t.status FROM tenant_user_links l JOIN tenants t ON t.id=l.tenant_id AND t.organization_id=l.organization_id WHERE l.organization_id=$1 AND l.user_id=$2 ORDER BY t.legal_name`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TenantProfile, 0)
	for rows.Next() {
		var item TenantProfile
		if err := rows.Scan(&item.ID, &item.LegalName, &item.Email, &item.Phone, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListTenantOccupancies(ctx context.Context, organizationID, userID string) ([]TenantOccupancy, error) {
	rows, err := r.pool.Query(ctx, `SELECT t.id, tn.id, tt.role, tn.status, tn.start_date::text, COALESCE(tn.end_date::text,''), p.id, p.name, u.id, u.label, COALESCE(l.id::text,''), COALESCE(l.reference_code,''), COALESCE(l.status,''), COALESCE(l.start_date::text,''), COALESCE(l.end_date::text,''), COALESCE(l.rent_amount_minor,0), COALESCE(l.deposit_amount_minor,0), COALESCE(l.currency,''), COALESCE(l.due_day,0) FROM tenant_user_links tul JOIN tenants t ON t.id=tul.tenant_id AND t.organization_id=tul.organization_id JOIN tenancy_tenants tt ON tt.organization_id=tul.organization_id AND tt.tenant_id=tul.tenant_id JOIN tenancies tn ON tn.id=tt.tenancy_id AND tn.organization_id=tt.organization_id JOIN units u ON u.id=tn.unit_id AND u.organization_id=tn.organization_id JOIN properties p ON p.id=u.property_id AND p.organization_id=u.organization_id LEFT JOIN LATERAL (SELECT lx.* FROM leases lx WHERE lx.organization_id=tn.organization_id AND lx.tenancy_id=tn.id ORDER BY CASE WHEN lx.status='active' THEN 0 ELSE 1 END, lx.end_date DESC LIMIT 1) l ON true WHERE tul.organization_id=$1 AND tul.user_id=$2 ORDER BY CASE tn.status WHEN 'active' THEN 0 WHEN 'upcoming' THEN 1 ELSE 2 END, tn.start_date DESC`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TenantOccupancy, 0)
	for rows.Next() {
		var item TenantOccupancy
		if err := rows.Scan(&item.TenantID, &item.TenancyID, &item.OccupantRole, &item.TenancyStatus, &item.StartDate, &item.EndDate, &item.PropertyID, &item.PropertyName, &item.UnitID, &item.UnitLabel, &item.LeaseID, &item.LeaseReference, &item.LeaseStatus, &item.LeaseStartDate, &item.LeaseEndDate, &item.RentAmountMinor, &item.DepositAmountMinor, &item.Currency, &item.DueDay); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListTenantRent(ctx context.Context, organizationID, userID string) ([]TenantRentItem, error) {
	rows, err := r.pool.Query(ctx, `WITH accessible AS (SELECT DISTINCT ON (tn.id) tn.id AS tenancy_id, tt.tenant_id FROM tenant_user_links tul JOIN tenancy_tenants tt ON tt.organization_id=tul.organization_id AND tt.tenant_id=tul.tenant_id JOIN tenancies tn ON tn.organization_id=tt.organization_id AND tn.id=tt.tenancy_id WHERE tul.organization_id=$1 AND tul.user_id=$2 ORDER BY tn.id, CASE WHEN tt.role='primary' THEN 0 ELSE 1 END) SELECT ro.id, a.tenant_id, a.tenancy_id, l.id, to_char(ro.period_start,'YYYY-MM'), ro.due_date::text, ro.amount_minor, COALESCE(pa.allocated_minor,0), GREATEST(ro.amount_minor-COALESCE(pa.allocated_minor,0),0), ro.currency, CASE WHEN ro.status='void' THEN 'void' WHEN ro.amount_minor-COALESCE(pa.allocated_minor,0)<=0 THEN 'paid' WHEN ro.due_date<CURRENT_DATE THEN 'overdue' ELSE 'open' END FROM accessible a JOIN leases l ON l.organization_id=$1 AND l.tenancy_id=a.tenancy_id JOIN rent_obligations ro ON ro.organization_id=l.organization_id AND ro.lease_id=l.id LEFT JOIN LATERAL (SELECT COALESCE(SUM(x.amount_minor),0)::bigint AS allocated_minor FROM payment_allocations x WHERE x.organization_id=ro.organization_id AND x.obligation_id=ro.id) pa ON true ORDER BY ro.due_date DESC`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TenantRentItem, 0)
	for rows.Next() {
		var item TenantRentItem
		if err := rows.Scan(&item.ObligationID, &item.TenantID, &item.TenancyID, &item.LeaseID, &item.Period, &item.DueDate, &item.AmountMinor, &item.AllocatedMinor, &item.BalanceMinor, &item.Currency, &item.State); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListTenantMaintenance(ctx context.Context, organizationID, userID string) ([]TenantMaintenanceItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT mr.id, mr.tenant_id, p.name, COALESCE(u.label,''), mr.title, mr.category, mr.priority, mr.status, mr.created_at, mr.updated_at FROM tenant_user_links tul JOIN maintenance_requests mr ON mr.organization_id=tul.organization_id AND mr.tenant_id=tul.tenant_id JOIN properties p ON p.organization_id=mr.organization_id AND p.id=mr.property_id LEFT JOIN units u ON u.organization_id=mr.organization_id AND u.id=mr.unit_id WHERE tul.organization_id=$1 AND tul.user_id=$2 ORDER BY mr.created_at DESC`, organizationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TenantMaintenanceItem, 0)
	for rows.Next() {
		var item TenantMaintenanceItem
		if err := rows.Scan(&item.ID, &item.TenantID, &item.PropertyName, &item.UnitLabel, &item.Title, &item.Category, &item.Priority, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetTenantMaintenanceContext(ctx context.Context, organizationID, userID, tenancyID string) (TenantMaintenanceContext, error) {
	var item TenantMaintenanceContext
	err := r.pool.QueryRow(ctx, `SELECT tt.tenant_id, u.property_id, tn.unit_id FROM tenant_user_links tul JOIN tenancy_tenants tt ON tt.organization_id=tul.organization_id AND tt.tenant_id=tul.tenant_id JOIN tenancies tn ON tn.organization_id=tt.organization_id AND tn.id=tt.tenancy_id JOIN units u ON u.organization_id=tn.organization_id AND u.id=tn.unit_id WHERE tul.organization_id=$1 AND tul.user_id=$2 AND tn.id=$3 AND tn.status='active' ORDER BY CASE WHEN tt.role='primary' THEN 0 ELSE 1 END LIMIT 1`, organizationID, userID, tenancyID).Scan(&item.TenantID, &item.PropertyID, &item.UnitID)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantMaintenanceContext{}, ErrTenancyNotAccessible
	}
	return item, err
}
