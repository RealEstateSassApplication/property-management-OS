package accounting

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) OwnerStatement(ctx context.Context, organizationID, ownerID string, from, to time.Time) (OwnerStatement, error) {
	var statement OwnerStatement
	statement.OwnerID = ownerID
	statement.From = from.Format("2006-01-02")
	statement.To = to.Format("2006-01-02")
	if err := r.pool.QueryRow(ctx, `SELECT legal_name FROM owners WHERE organization_id=$1 AND id=$2 AND status='active'`, organizationID, ownerID).Scan(&statement.OwnerName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OwnerStatement{}, ErrOwnerNotFound
		}
		return OwnerStatement{}, err
	}

	lines := make([]StatementLine, 0)
	incomeRows, err := r.pool.Query(ctx, `
		SELECT to_char(pmt.received_at,'YYYY-MM-DD'),pr.id,pr.name,
		       'Rent collection',pmt.currency,pa.amount_minor,oi.ownership_bps,
		       (pa.amount_minor * oi.ownership_bps / 10000)::bigint
		FROM payment_allocations pa
		JOIN payments pmt ON pmt.id=pa.payment_id AND pmt.organization_id=pa.organization_id AND pmt.status='posted'
		JOIN rent_obligations ro ON ro.id=pa.obligation_id AND ro.organization_id=pa.organization_id
		JOIN leases l ON l.id=ro.lease_id AND l.organization_id=ro.organization_id
		JOIN tenancies tn ON tn.id=l.tenancy_id AND tn.organization_id=l.organization_id
		JOIN units u ON u.id=tn.unit_id AND u.organization_id=tn.organization_id
		JOIN properties pr ON pr.id=u.property_id AND pr.organization_id=u.organization_id
		JOIN ownership_interests oi ON oi.organization_id=pr.organization_id AND oi.property_id=pr.id AND oi.owner_id=$2
		WHERE pa.organization_id=$1 AND pmt.received_at BETWEEN $3::date AND $4::date
		  AND oi.effective_from<=pmt.received_at
		  AND (oi.effective_to IS NULL OR oi.effective_to>=pmt.received_at)
		ORDER BY pmt.received_at,pa.created_at
	`, organizationID, ownerID, from, to)
	if err != nil {
		return OwnerStatement{}, err
	}
	for incomeRows.Next() {
		var line StatementLine
		line.LineType = "income"
		if err := incomeRows.Scan(&line.Date, &line.PropertyID, &line.PropertyName, &line.Description, &line.Currency, &line.GrossMinor, &line.OwnershipBPS, &line.OwnerMinor); err != nil {
			incomeRows.Close()
			return OwnerStatement{}, err
		}
		lines = append(lines, line)
	}
	if err := incomeRows.Err(); err != nil {
		incomeRows.Close()
		return OwnerStatement{}, err
	}
	incomeRows.Close()

	expenseRows, err := r.pool.Query(ctx, `
		SELECT to_char(e.incurred_on,'YYYY-MM-DD'),pr.id,pr.name,
		       e.category || ': ' || e.note,e.currency,e.amount_minor,oi.ownership_bps,
		       (e.amount_minor * oi.ownership_bps / 10000)::bigint
		FROM property_expenses e
		JOIN properties pr ON pr.id=e.property_id AND pr.organization_id=e.organization_id
		JOIN ownership_interests oi ON oi.organization_id=pr.organization_id AND oi.property_id=pr.id AND oi.owner_id=$2
		LEFT JOIN property_expense_reversals er ON er.organization_id=e.organization_id AND er.expense_id=e.id
		WHERE e.organization_id=$1 AND e.incurred_on BETWEEN $3::date AND $4::date
		  AND er.id IS NULL
		  AND oi.effective_from<=e.incurred_on
		  AND (oi.effective_to IS NULL OR oi.effective_to>=e.incurred_on)
		ORDER BY e.incurred_on,e.created_at
	`, organizationID, ownerID, from, to)
	if err != nil {
		return OwnerStatement{}, err
	}
	for expenseRows.Next() {
		var line StatementLine
		line.LineType = "expense"
		if err := expenseRows.Scan(&line.Date, &line.PropertyID, &line.PropertyName, &line.Description, &line.Currency, &line.GrossMinor, &line.OwnershipBPS, &line.OwnerMinor); err != nil {
			expenseRows.Close()
			return OwnerStatement{}, err
		}
		lines = append(lines, line)
	}
	if err := expenseRows.Err(); err != nil {
		expenseRows.Close()
		return OwnerStatement{}, err
	}
	expenseRows.Close()

	sort.SliceStable(lines, func(i, j int) bool {
		if lines[i].Date == lines[j].Date {
			return lines[i].LineType < lines[j].LineType
		}
		return lines[i].Date < lines[j].Date
	})
	statement.Lines = lines

	summaryMap := make(map[string]*StatementCurrencySummary)
	for _, line := range lines {
		summary := summaryMap[line.Currency]
		if summary == nil {
			summary = &StatementCurrencySummary{Currency: line.Currency}
			summaryMap[line.Currency] = summary
		}
		if line.LineType == "income" {
			summary.IncomeMinor += line.OwnerMinor
		} else {
			summary.ExpenseMinor += line.OwnerMinor
		}
	}
	currencies := make([]string, 0, len(summaryMap))
	for currency := range summaryMap {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	statement.Summaries = make([]StatementCurrencySummary, 0, len(currencies))
	for _, currency := range currencies {
		summary := *summaryMap[currency]
		summary.NetOwnerAmountMinor = summary.IncomeMinor - summary.ExpenseMinor
		statement.Summaries = append(statement.Summaries, summary)
	}
	return statement, nil
}
