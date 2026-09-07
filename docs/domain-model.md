# Domain Model

## Core aggregate flow

```text
Organization
  |
  +-- Memberships -- Users
  |
  +-- Properties
        |
        +-- Units
              |
              +-- Tenancies
                    |
                    +-- Tenancy tenants -- Tenants
                    |
                    +-- Leases
                          |
                          +-- Rent obligations
                          +-- Payments / allocations

Property / Unit / Tenancy
  |
  +-- Maintenance requests
        +-- Work orders
              +-- Vendors
```

## Organization and authorization

The SaaS tenant. A property-management company, landlord business, or other portfolio operator is represented as an organization.

Users gain access through `organization_memberships`. Membership role is verified server-side on each business request. The current generalized permissions are:

- `admin`, `manager`: portfolio, people, and leasing read/write
- `accountant`, `viewer`: read-only portfolio/people/leasing views
- `maintenance`: portfolio read access
- `owner`: no generalized organization access until owner-resource scoping is implemented

Development identity headers are only an adapter for local identity injection; they do not bypass membership checks.

## Property

A physical managed asset such as a house, apartment building, commercial property, or other real-estate asset. A property can contain one or more units.

## Unit

The rentable/occupiable entity. A single-family house may still be represented with a single unit so leasing and occupancy logic stays consistent.

Only one tenancy may be in `active` state for a unit at a time. PostgreSQL enforces this invariant with a partial unique index.

## Owner

A person or legal entity with an ownership interest in one or more properties. Ownership percentages and effective dates remain a separate upcoming domain from organization membership and portal access.

## Tenant

A person or legal entity entering a tenancy. Tenant identity is distinct from a portal login so a tenant record can exist before an account is invited and remain historically valid if login identities change.

Current lifecycle:

`prospect -> active -> former`

`blocked` is an exceptional state that prevents creating new tenancy relationships.

## Tenancy

Represents the occupancy/business relationship for a unit. It connects one or more tenants to a unit and provides continuity across lease renewals.

A tenancy has exactly one primary tenant and can have additional occupants through `tenancy_tenants`.

Current lifecycle:

`upcoming -> active -> ended`

`cancelled` is an exceptional terminal state. Activating a tenancy marks the unit occupied. Ending/cancelling an active tenancy releases the unit to vacant state.

## Lease

The contractual terms for a tenancy over a defined period. A tenancy can have multiple historical leases, but only one lease can be active at once.

Current lifecycle:

`draft -> active -> expired`

Exceptional paths:

- `draft -> cancelled`
- `active -> terminated`

Final states are immutable through the normal status-transition service.

Lease rent and deposit amounts are stored as integer **minor units** (`BIGINT`), never floating-point values. For example, LKR 150,000.00 is stored as `15000000` minor units. This keeps future rent obligations, allocations, credits, and arrears deterministic.

## Rent ledger

Rent will not be represented only by fields such as `monthly_rent` and `paid`. The financial model will use obligations, transactions, allocations, credits, and adjustments so arrears and historical balances can be reconstructed.

Example:

```text
Rent obligation: 10,000,000 minor units
Payment:          6,000,000 minor units
Allocation:       6,000,000 -> obligation
Outstanding:      4,000,000 minor units
```

The Go backend will be authoritative for balances and status.

## Maintenance

Maintenance begins as a request and may produce one or more work orders. The system should retain issue history, assignment, quotes/approvals, scheduling, status transitions, costs, attachments, and proof of completion.

## Documents

Documents attach to explicit resources and have ownership/access metadata. Examples include leases, IDs, inspection reports, invoices, receipts, and maintenance evidence.

## Audit

Security-sensitive and financially relevant actions produce immutable audit events. Audit is not a substitute for domain history; both are retained where appropriate.

## Implementation sequence

Completed foundation:

1. Organizations, users, membership schema
2. Properties and units
3. Membership/RBAC authorization boundary
4. Tenants and tenancies
5. Leases

Next:

6. Owners and ownership interests
7. Rent obligations and ledger
8. Payments and allocations
9. Maintenance and vendors
10. Documents and notifications
11. Owner/tenant portal resource scoping
12. Reporting and automation/integration workflows
