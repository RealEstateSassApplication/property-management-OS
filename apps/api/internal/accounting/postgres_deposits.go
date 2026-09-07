package accounting

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const depositAccountSelect = `
	SELECT da.id,da.organization_id,da.lease_id,l.reference_code,t.legal_name,p.name,u.label,
	       da.required_amount_minor,
	       COALESCE(SUM(CASE WHEN dt.transaction_type IN ('received','adjustment_increase') THEN dt.amount_minor ELSE -dt.amount_minor END),0)::bigint,
	       da.currency,da.created_at
	FROM security_deposit_accounts da
	JOIN leases l ON l.id=da.lease_id AND l.organization_id=da.organization_id
	JOIN tenancies tn ON tn.id=l.tenancy_id AND tn.organization_id=l.organization_id
	JOIN units u ON u.id=tn.unit_id AND u.organization_id=tn.organization_id
	JOIN properties p ON p.id=u.property_id AND p.organization_id=u.organization_id
	JOIN tenancy_tenants tt ON tt.tenancy_id=tn.id AND tt.organization_id=tn.organization_id AND tt.role='primary'
	JOIN tenants t ON t.id=tt.tenant_id AND t.organization_id=tt.organization_id
	LEFT JOIN security_deposit_transactions dt ON dt.deposit_account_id=da.id AND dt.organization_id=da.organization_id
`

func (r *PostgresRepository) ListDepositAccounts(ctx context.Context, organizationID string) ([]DepositAccount, error) {
	rows, err := r.pool.Query(ctx, depositAccountSelect+`
		WHERE da.organization_id=$1
		GROUP BY da.id,l.reference_code,t.legal_name,p.name,u.label
		ORDER BY da.created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]DepositAccount, 0)
	for rows.Next() {
		var item DepositAccount
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.LeaseID, &item.LeaseReference, &item.TenantName, &item.PropertyName, &item.UnitLabel, &item.RequiredAmountMinor, &item.HeldAmountMinor, &item.Currency, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateDepositAccount(ctx context.Context, organizationID string, input CreateDepositAccountInput) (DepositAccount, error) {
	var required int64
	var currency string
	if err := r.pool.QueryRow(ctx, `SELECT deposit_amount_minor,currency FROM leases WHERE organization_id=$1 AND id=$2`, organizationID, input.LeaseID).Scan(&required, &currency); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DepositAccount{}, ErrLeaseNotFound
		}
		return DepositAccount{}, err
	}
	var accountID string
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO security_deposit_accounts (organization_id,lease_id,required_amount_minor,currency)
		VALUES ($1,$2,$3,$4)
		RETURNING id
	`, organizationID, input.LeaseID, required, currency).Scan(&accountID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return DepositAccount{}, ErrDepositAccountExists
		}
		return DepositAccount{}, err
	}
	return r.getDepositAccount(ctx, organizationID, accountID)
}

func (r *PostgresRepository) getDepositAccount(ctx context.Context, organizationID, accountID string) (DepositAccount, error) {
	var item DepositAccount
	err := r.pool.QueryRow(ctx, depositAccountSelect+`
		WHERE da.organization_id=$1 AND da.id=$2
		GROUP BY da.id,l.reference_code,t.legal_name,p.name,u.label
	`, organizationID, accountID).Scan(
		&item.ID, &item.OrganizationID, &item.LeaseID, &item.LeaseReference, &item.TenantName, &item.PropertyName,
		&item.UnitLabel, &item.RequiredAmountMinor, &item.HeldAmountMinor, &item.Currency, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DepositAccount{}, ErrDepositAccountNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListDepositTransactions(ctx context.Context, organizationID string) ([]DepositTransaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,organization_id,deposit_account_id,transaction_type,amount_minor,to_char(occurred_on,'YYYY-MM-DD'),note,
		       COALESCE(created_by_user_id::text,''),created_at
		FROM security_deposit_transactions
		WHERE organization_id=$1
		ORDER BY occurred_on DESC,created_at DESC
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]DepositTransaction, 0)
	for rows.Next() {
		var item DepositTransaction
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.DepositAccountID, &item.TransactionType, &item.AmountMinor, &item.OccurredOn, &item.Note, &item.CreatedByUserID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateDepositTransaction(ctx context.Context, organizationID, actorUserID string, input CreateDepositTransactionInput) (DepositTransaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return DepositTransaction{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var accountID string
	if err := tx.QueryRow(ctx, `
		SELECT id::text FROM security_deposit_accounts
		WHERE organization_id=$1 AND id=$2
		FOR UPDATE
	`, organizationID, input.DepositAccountID).Scan(&accountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DepositTransaction{}, ErrDepositAccountNotFound
		}
		return DepositTransaction{}, err
	}

	var held int64
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE WHEN transaction_type IN ('received','adjustment_increase') THEN amount_minor ELSE -amount_minor END),0)::bigint
		FROM security_deposit_transactions
		WHERE organization_id=$1 AND deposit_account_id=$2
	`, organizationID, input.DepositAccountID).Scan(&held); err != nil {
		return DepositTransaction{}, err
	}
	if input.TransactionType == "deduction" || input.TransactionType == "refund" || input.TransactionType == "adjustment_decrease" {
		if input.AmountMinor > held {
			return DepositTransaction{}, ErrDepositInsufficientFunds
		}
	}

	var item DepositTransaction
	if err := tx.QueryRow(ctx, `
		INSERT INTO security_deposit_transactions (organization_id,deposit_account_id,transaction_type,amount_minor,occurred_on,note,created_by_user_id)
		VALUES ($1,$2,$3,$4,$5::date,$6,NULLIF($7,'')::uuid)
		RETURNING id,organization_id,deposit_account_id,transaction_type,amount_minor,to_char(occurred_on,'YYYY-MM-DD'),note,COALESCE(created_by_user_id::text,''),created_at
	`, organizationID, input.DepositAccountID, input.TransactionType, input.AmountMinor, input.OccurredOn, input.Note, actorUserID).Scan(
		&item.ID, &item.OrganizationID, &item.DepositAccountID, &item.TransactionType, &item.AmountMinor, &item.OccurredOn, &item.Note, &item.CreatedByUserID, &item.CreatedAt,
	); err != nil {
		return DepositTransaction{}, err
	}
	if err := insertAudit(ctx, tx, organizationID, actorUserID, "accounting.deposit.transaction_created", "security_deposit_account", input.DepositAccountID, fmt.Sprintf(`{"transactionId":"%s","type":"%s","amountMinor":%d}`, item.ID, item.TransactionType, item.AmountMinor)); err != nil {
		return DepositTransaction{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DepositTransaction{}, err
	}
	return item, nil
}
