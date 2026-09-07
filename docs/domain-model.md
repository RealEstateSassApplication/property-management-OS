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
                    +-- Tenants
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

## Organization

The SaaS tenant. A property management company, landlord business, or other portfolio operator is represented as an organization.

## Property

A physical managed asset such as a house, apartment building, commercial property, or other real-estate asset. A property can contain one or more units.

## Unit

The rentable/occupiable entity. A single-family house may still be represented with a single unit so leasing and occupancy logic stays consistent.

## Owner

A person or legal entity with an ownership interest in one or more properties. Ownership percentages and effective dates will be modeled separately from user access.

## Tenant

A person or legal entity entering a tenancy. Tenant identity is distinct from a portal login so a tenant record can exist before an account is invited.

## Tenancy

Represents occupancy/business relationship for a unit. It connects one or more tenants to a unit and provides continuity across lease amendments or renewals.

## Lease

The contractual terms for a tenancy over a defined period. Lease state transitions must be explicit and auditable.

Suggested lifecycle:

`draft -> pending_signature -> active -> ended`

Additional exceptional states such as cancelled or terminated should preserve historical records rather than deleting them.

## Rent ledger

Rent should not be represented only by fields such as `monthly_rent` and `paid`. The financial model will use obligations, transactions, allocations, credits, and adjustments so arrears and historical balances can be reconstructed.

Example:

```text
Rent obligation: LKR 100,000
Payment:         LKR  60,000
Allocation:      LKR  60,000 -> obligation
Outstanding:     LKR  40,000
```

The backend is authoritative for balances and status.

## Maintenance

Maintenance begins as a request and may produce one or more work orders. The system should retain issue history, assignment, quotes/approvals, scheduling, status transitions, costs, attachments, and proof of completion.

## Documents

Documents attach to explicit resources and have ownership/access metadata. Examples include leases, IDs, inspection reports, invoices, receipts, and maintenance evidence.

## Audit

Security-sensitive and financially relevant actions produce immutable audit events. Audit is not a substitute for domain history; both are retained where appropriate.

## Planned implementation sequence

1. Organizations, users, memberships, RBAC
2. Properties and units
3. Owners and ownership interests
4. Tenants and tenancies
5. Leases
6. Rent obligations and ledger
7. Payments and allocations
8. Maintenance and vendors
9. Documents
10. Notifications
11. Reporting
12. Automation/integration workflows
