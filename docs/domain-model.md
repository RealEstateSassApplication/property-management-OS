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
  |
  +-- Vendors
  |
  +-- Maintenance requests
        |
        +-- Work orders
              |
              +-- Maintenance quotes
              |
              +-- Completion evidence
```

## Organization and authorization

The SaaS tenant. A property-management company, landlord business, or other portfolio operator is represented as an organization.

Users gain access through `organization_memberships`. Membership role is verified server-side on every business request. Current generalized permissions are:

- `admin`, `manager`: read/write across implemented organization modules
- `accountant`: read portfolio/people/leasing/owners/maintenance, manage rent, and approve maintenance costs
- `viewer`: read-only portfolio/people/leasing/owners/rent/maintenance
- `maintenance`: portfolio read plus maintenance execution; cannot manage vendor master data or approve costs
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

Obligation financial state is derived from obligation and allocation rows rather than stored as a mutable paid flag.

## Payment and allocation

A payment is an immutable posted cash receipt tied to a tenant. A payment allocation applies part of that receipt to a rent obligation. Row locking, currency validation, tenancy validation, and remaining-balance checks protect the ledger from over-allocation and concurrent double spend.

## Vendor

A vendor is an organization-scoped service provider used by maintenance operations. Vendors have a trade and lifecycle state. An inactive vendor cannot be assigned to new work or submit new quotes.

Vendor master-data writes are intentionally separated from ordinary maintenance execution permissions.

## Maintenance request

A maintenance request is the intake record for an operational issue. It belongs to a property and may be linked to a unit and tenant.

If a tenant is attached, the backend verifies that the tenant is an active occupant of the selected unit. This prevents arbitrary tenant/unit associations and keeps tenant-facing history trustworthy.

Lifecycle:

```text
open -> triaged -> in_progress -> resolved
  \        \             \
   +--------+-------------+-> cancelled
```

A resolved request must have at least one completed work order and no incomplete work orders.

## Work order

A work order is the executable job created from a maintenance request. It can be internal or assigned to a vendor and may be scheduled.

Lifecycle:

```text
planned -> assigned -> in_progress -> completed
   \          \             \
    +----------+-------------+-> cancelled
```

Starting a work order pushes its request into `in_progress`. Completing the final active work order resolves the request automatically when at least one work order completed successfully.

A work order **cannot be completed without completion evidence**. This is enforced transactionally in the backend, not as a UI convention.

## Maintenance quote and approval

A quote captures vendor scope and money in integer minor units.

Lifecycle:

`submitted -> approved | rejected | withdrawn`

Important controls:

- only one approved quote is allowed per work order
- quote approval requires the dedicated `maintenance:approve_costs` permission
- a maintenance operator cannot approve the quote they are executing under the generalized maintenance role
- approving a quote atomically assigns the selected vendor to the work order
- competing submitted quotes are rejected when one quote is approved
- decision actor and time are retained, and the action is also written to `audit_events`

This keeps cost approval separate from job execution while avoiding a separate workflow service at the current scale.

## Completion evidence

Evidence is immutable proof attached to a work order. Today it can be a note or a storage reference; the model already supports future photos, invoices, receipts, and other stored artifacts.

Each evidence record retains the submitting user and timestamp. Evidence creation also writes an audit event.

Example maintenance flow:

```text
Tenant-linked request
  -> triage
  -> work order
  -> vendor quote
  -> independent quote approval
  -> work starts
  -> completion evidence
  -> work order completes
  -> request resolves
```

## Documents

Documents will attach to explicit resources with ownership/access metadata. Examples include leases, IDs, inspection reports, invoices, receipts, and richer maintenance evidence.

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
9. Maintenance requests, vendors, work orders, quotes, approvals, and completion evidence

Next:

10. Production OIDC/JWT adapter
11. Documents, object storage, and notifications
12. Owner/tenant portal resource scoping
13. Financial adjustments, reversals, deposits, owner statements, reconciliation
14. Maintenance invoices/expenses and owner-specific approval policies
15. Reporting and automation/integration workflows
16. Avara/Blu integration adapters
