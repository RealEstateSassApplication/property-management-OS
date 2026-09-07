# Accounting controls

Property Management OS treats accounting changes as append-only operational records rather than editable balances.

## Rent assessments

Each rent obligation keeps `base_amount_minor` as the original lease-derived assessment. The current `amount_minor` remains the effective assessment used by existing rent, portal and reporting queries. Immutable `rent_adjustments` explain every change:

- `charge` and `late_fee` increase the effective assessment.
- `credit` and `writeoff` reduce it but cannot reduce the assessment below already allocated cash.
- `payment_reversal` is system-generated when a posted payment is reversed.

This preserves the original assessment while keeping one authoritative receivable calculation.

## Payment reversals

Payments and allocations are never deleted. Reversal locks the payment, creates one `payment_reversals` row, creates compensating `payment_reversal` adjustments for every affected obligation, and marks the original payment `void`. The old allocations remain available for audit while the outstanding receivable reopens.

## Security deposits

Security deposits are liabilities, not rental income. The account copies its required amount and currency from the lease. Held funds are derived from an append-only transaction ledger:

`received + adjustment_increase - deduction - refund - adjustment_decrease`

The deposit account row is locked before any deduction/refund/decrease balance check, preventing concurrent requests from overspending held funds.

## Property expenses

Expenses belong to an organization and property and can optionally reference a vendor and work order. Reversals are separate records; the original expense remains immutable. Owner statements exclude reversed expenses.

## Owner statements

Owner statements are derived for an explicit date range. Rent income uses posted payment allocations, while expenses use unreversed property expenses. Each line applies the ownership interest effective on the transaction date:

`owner amount = gross amount × ownership basis points / 10,000`

This means ownership changes do not retroactively rewrite historic owner economics.

## Authorization

- `admin`: full accounting access.
- `accountant`: view and manage accounting.
- `manager`: view accounting, but no accounting mutations.
- `viewer`, `maintenance`, `agent`, `owner`, and `tenant`: no general accounting access.

Owner-facing financial data continues through the resource-scoped owner portal rather than the manager accounting API.
