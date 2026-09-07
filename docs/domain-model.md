# Domain Model

## Core aggregate flow

```text
Organization
  |
  +-- Memberships -- Users
  |
  +-- Owners -- Ownership interests -- Properties
  |                                  |
  |                                  +-- Units
  |                                        |
  |                                        +-- Tenancies
  |                                              |
  |                                              +-- Tenancy tenants -- Tenants
  |                                              |
  |                                              +-- Leases
  |                                                    |
  |                                                    +-- Rent obligations
  |                                                          ^
  |                                                          |
  |                                                    Payment allocations
  |                                                          |
  +------------------------------------------------------ Payments

Property / Unit / Tenancy
  |
  +-- Maintenance requests
        +-- Work orders
              +-- Vendors
```

## Organization and authorization

The SaaS tenant. A property-management company, landlord business, or other portfolio operator is represented as an organization.

Users gain access through `organization_memberships`. Membership role is verified server-side on every business request. Current generalized permissions are:

- `admin`, `manager`: read/write across implemented organization modules
- `accountant`: read portfolio/people/leasing/owners and manage rent operations
- `viewer`: read-only portfolio/people/leasing/owners/rent
- `maintenance`: portfolio read access
- `owner`: no generalized organization access until owned-property resource scoping is implemented

Development identity headers only inject local identity; they do not bypass membership or role checks.

## Property and unit

A property is a physical managed asset. A unit is the rentable/occupiable entity within that asset. Single-family assets can still use one unit so occupancy and lease rules remain consistent.

Only one tenancy may be `active` for a unit at a time. PostgreSQL enforces this with a partial unique index.

## Owner and ownership interest

An owner is a person or legal entity with economic ownership in a property. It is deliberately separate from an application user or organization membership.

An `ownership_interest` links an owner to a property using **basis points**:

```text
10,000 bps = 100.00%
 5,000 bps =  50.00%
    25 bps =   0.25%
```

Interests have effective dates so historical ownership can be preserved. Creating a current interest locks the property row, sums current open interests, and rejects any change that would push total current ownership above 100%.

Portal access for an owner will later be resource-scoped to the properties that owner is authorized to see; ownership data alone does not automatically grant login access.

## Tenant

A person or legal entity entering a tenancy. Tenant identity is distinct from portal login identity so a tenant can exist before receiving an account and remain historically valid if authentication identities change.

Current lifecycle:

`prospect -> active -> former`

`blocked` is an exceptional state that prevents new tenancy relationships.

## Tenancy

The occupancy/business relationship for a unit. It connects one or more tenants to a unit and remains continuous across lease renewals.

A tenancy has exactly one primary tenant and may have additional occupants through `tenancy_tenants`.

Current lifecycle:

`upcoming -> active -> ended`

`cancelled` is an exceptional terminal state. Activating a tenancy marks the unit occupied. Ending/cancelling an active tenancy releases the unit.

## Lease

The contractual terms for a tenancy over a defined period. A tenancy may have multiple historical leases, but only one lease can be active at once.

Current lifecycle:

`draft -> active -> expired`

Exceptional paths:

- `draft -> cancelled`
- `active -> terminated`

Lease rent and deposit amounts are stored as integer **minor units** (`BIGINT`), never floating point. LKR 150,000.00 is stored as `15000000` minor units.

## Rent obligation

A rent obligation represents one lease-backed receivable for one month. The client supplies only the active lease and period; the backend derives amount, currency, and due day from the lease.

There can be only one obligation per organization/lease/month. The due day is capped to the last valid day of short months.

Obligation financial state is not stored as `paid=true`. It is derived:

```text
allocated = SUM(payment_allocations.amount_minor)
balance   = obligation.amount_minor - allocated

void      -> obligation record explicitly voided
paid      -> allocated >= amount
overdue   -> balance > 0 and current date > due date
open      -> balance > 0 and not overdue
```

## Payment

A payment is an immutable posted cash receipt tied to a tenant. It records amount, currency, received date, method, and optional external reference.

The unallocated payment balance is derived from allocations:

```text
unallocated = payment.amount_minor - SUM(payment_allocations.amount_minor)
```

Future reversals should be explicit accounting events, not silent edits to a posted payment.

## Payment allocation

An allocation applies some or all of a payment to a rent obligation. Multiple allocations can be created, allowing partial payments and incremental allocation.

During allocation the backend locks both the payment and obligation rows and verifies:

- payment is posted
- obligation is not void
- currencies match
- payment tenant belongs to the obligation tenancy
- allocation does not exceed remaining payment balance
- allocation does not exceed remaining obligation balance

This protects the ledger from concurrent double allocation and overpayment races.

Example:

```text
Rent obligation: LKR 150,000
Payment:         LKR 100,000
Allocation:      LKR 100,000
Outstanding:     LKR  50,000
```

## Maintenance

Maintenance begins as a request and may produce one or more work orders. The system should retain issue history, assignment, quotes/approvals, scheduling, status transitions, costs, attachments, and proof of completion.

## Documents

Documents attach to explicit resources and have ownership/access metadata. Examples include leases, IDs, inspection reports, invoices, receipts, and maintenance evidence.

## Audit

Security-sensitive and financially relevant actions produce immutable audit events. Audit is not a substitute for domain history; both are retained where appropriate.

## Implementation sequence

Completed:

1. Organizations, users, membership schema
2. Properties and units
3. Membership/RBAC authorization boundary
4. Tenants and tenancies
5. Leases
6. Owners and ownership interests
7. Rent obligations
8. Payments and allocations

Next:

9. Maintenance requests, vendors, work orders, quotes, approvals
10. Production OIDC/JWT adapter
11. Documents and notifications
12. Owner/tenant portal resource scoping
13. Financial adjustments, reversals, deposits, owner statements, reconciliation
14. Reporting and automation/integration workflows
