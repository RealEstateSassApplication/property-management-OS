# Domain Model

## Core aggregate flow

```text
OIDC issuer + subject
        |
        v
User identities -> Users -> Memberships -> Organization
                                         |
                                         +-- Owners -- Ownership interests -- Properties
                                         |                                  |
                                         |                                  +-- Units
                                         |                                       |
                                         |                                       +-- Tenancies
                                         |                                             |
                                         |                                             +-- Tenancy tenants -- Tenants
                                         |                                             |
                                         |                                             +-- Leases -> Rent obligations
                                         |                                                            ^
                                         |                                                            |
                                         |                                                       Allocations
                                         |                                                            |
                                         +-------------------------------------------------------- Payments
                                         |
                                         +-- Vendors
                                         |
                                         +-- Maintenance requests -> Work orders -> Quotes / evidence
```

## Organization

The SaaS tenant. A property-management company, landlord business, or other portfolio operator is represented as an organization. All business records are organization-scoped.

`X-Organization-ID` is an explicit requested organization scope. Possessing or guessing an organization ID does not grant access; membership and route permission are verified after authentication.

## User identity and authentication

`users` are stable Property OS identities. External login-provider identifiers are deliberately stored separately in `user_identities`.

A production request follows this chain:

```text
Authorization: Bearer <JWT>
        ↓
OIDC discovery / JWKS signature / issuer / audience / lifetime verification
        ↓
verified issuer + subject
        ↓
user_identities (issuer, subject)
        ↓
internal users.id
        ↓
organization_memberships
        ↓
role permission
```

Important identity invariants:

- OIDC `sub` is never treated as an internal UUID.
- `(issuer, subject)` is unique and is the external identity key.
- A single internal user can later be linked to multiple external identities/providers without changing domain references.
- Token email is metadata only (`last_seen_email`); it never automatically links or authorizes a user.
- Invalid, expired, wrong-audience, wrong-issuer, or unverifiable tokens are authentication failures (`401`).
- A cryptographically valid identity with no `user_identities` mapping is authenticated externally but not provisioned locally (`403`).
- Production startup requires `OIDC_ISSUER_URL` and `OIDC_AUDIENCE`.
- Development/test may inject an internal user with `X-User-ID` when no Bearer token is present. Bearer authentication takes precedence when supplied.

## Authorization

Users gain organization access through `organization_memberships`. Membership role is verified server-side on every business request after identity has been resolved to an internal user.

Current generalized permissions are:

- `admin`, `manager`: read/write across implemented organization modules
- `accountant`: read portfolio/people/leasing/owners/maintenance, manage rent, and approve maintenance costs
- `viewer`: read-only portfolio/people/leasing/owners/rent/maintenance
- `maintenance`: portfolio read plus maintenance execution; cannot manage vendor master data or approve costs
- `owner`: no generalized organization access until owned-property resource scoping is implemented

Authentication and authorization are intentionally separate: a valid login never implies access to a particular organization.

## Property and unit

A property is a physical managed asset. A unit is the rentable/occupiable entity within that asset. Single-family assets can still use one unit so occupancy and lease rules remain consistent.

Only one tenancy may be `active` for a unit at a time. PostgreSQL enforces this with a partial unique index.

## Owner and ownership interest

An owner is a person or legal entity with economic ownership in a property. It is deliberately separate from an application user or organization membership.

An `ownership_interest` links an owner to a property using basis points:

```text
10,000 bps = 100.00%
 5,000 bps =  50.00%
    25 bps =   0.25%
```

Interests have effective dates so historical ownership can be preserved. Creating a current interest locks the property row, sums current open interests, and rejects any change that would push total current ownership above 100%.

Ownership does not automatically create portal authorization; owner portal access will be resource-scoped explicitly.

## Tenant

A person or legal entity entering a tenancy. Tenant identity is distinct from portal login identity so a tenant can exist before receiving an account and remain historically valid if authentication identities change.

Current lifecycle:

`prospect -> active -> former`

`blocked` is an exceptional state that prevents new tenancy relationships.

## Tenancy

The occupancy/business relationship for a unit. It connects one or more tenants to a unit and remains continuous across lease renewals.

A tenancy has exactly one primary tenant and may have additional occupants through `tenancy_tenants`.

Lifecycle:

`upcoming -> active -> ended`

`cancelled` is an exceptional terminal state. Activating a tenancy marks the unit occupied. Ending/cancelling an active tenancy releases the unit.

## Lease

The contractual terms for a tenancy over a defined period. A tenancy may have multiple historical leases, but only one lease can be active at once.

Lifecycle:

`draft -> active -> expired`

Exceptional paths:

- `draft -> cancelled`
- `active -> terminated`

Rent and deposit amounts are stored as integer minor units (`BIGINT`), never floating point.

## Rent obligation

A rent obligation represents one lease-backed receivable for one month. The client supplies only the active lease and period; the backend derives amount, currency, and due day from the lease.

There can be only one obligation per organization/lease/month. Financial state is derived from obligation and allocation rows rather than stored as a mutable paid flag.

## Payment and allocation

A payment is an immutable posted cash receipt tied to a tenant. An allocation applies part of that receipt to a rent obligation. Row locking, currency validation, tenancy validation, and remaining-balance checks protect the ledger from over-allocation and concurrent double spend.

## Vendor

A vendor is an organization-scoped service provider used by maintenance operations. Vendors have a trade and lifecycle state. An inactive vendor cannot be assigned to new work or submit new quotes.

Vendor master-data writes are separated from ordinary maintenance execution permissions.

## Maintenance request

A maintenance request is the intake record for an operational issue. It belongs to a property and may be linked to a unit and tenant.

If a tenant is attached, the backend verifies that the tenant is an active occupant of the selected unit.

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

A work order cannot be completed without completion evidence. This is enforced transactionally in the backend.

## Maintenance quote and approval

A quote captures vendor scope and money in integer minor units.

Lifecycle:

`submitted -> approved | rejected | withdrawn`

Controls:

- only one approved quote per work order
- approval requires `maintenance:approve_costs`
- ordinary maintenance execution does not imply cost-approval permission
- approval atomically assigns the selected vendor
- competing submitted quotes are rejected when one is approved
- decision actor/time and audit event are retained

## Completion evidence

Evidence is immutable proof attached to a work order. It can be a note or storage reference and is designed to support future photos, invoices, receipts, and other artifacts.

Each evidence record retains submitting user and timestamp.

## Documents

Documents will attach to explicit resources with ownership/access metadata. Files themselves should live in object storage rather than PostgreSQL; PostgreSQL should store controlled metadata and relationships.

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
10. Production OIDC/JWT authentication and external-to-internal user identity mapping

Next:

11. Documents, object storage, and notifications
12. Owner/tenant portal resource scoping
13. Financial adjustments, reversals, deposits, owner statements, reconciliation
14. Maintenance invoices/expenses and owner-specific approval policies
15. Reporting and automation/integration workflows
16. Avara/Blu integration adapters
