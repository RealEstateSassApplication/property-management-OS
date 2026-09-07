# Property Management OS

Property Management OS is the operational layer for managing rental properties, units, owners, tenants, occupancies, leases, rent, maintenance, documents, vendors, and reporting.

## Architecture

This repository is a monorepo built around a deliberately simple startup architecture:

- **Web:** Next.js + TypeScript
- **Backend:** Go modular monolith
- **Database:** PostgreSQL
- **API contract:** OpenAPI
- **Local development:** Docker Compose

The system is multi-tenant from day one. Business data is organization-scoped and authorization is enforced server-side.

## Repository layout

```text
apps/
  web/                 Next.js property manager, owner, and tenant experiences
  api/                 Go API and future background-worker entrypoints
contracts/
  openapi.yaml         Source API contract
migrations/            PostgreSQL migrations
scripts/               Local development helpers and seed data
docs/                  Architecture and domain documentation
```

## Implemented

### Foundation and portfolio

- Next.js application shell
- Go HTTP API with graceful shutdown and structured logging
- PostgreSQL connection pooling with pgx
- organization-scoped properties and units
- property and unit create/list/get/update APIs
- portfolio register, property details, unit management, and occupancy metrics
- OpenAPI contract and CI for Go, PostgreSQL, and Next.js

### Identity and authorization

- development identity includes organization and user context
- every business request verifies `organization_memberships`
- route-level permissions for portfolio, people, leasing, owners, rent, and maintenance
- admin/manager can manage all current organization modules
- accountant can read operational data, manage rent, and approve maintenance costs
- viewer is read-only
- maintenance role can view portfolio and execute maintenance operations but cannot approve vendor costs or manage vendor master data
- owner role intentionally has no generalized organization access until resource-scoped owner portal authorization exists
- development identity headers are disabled outside development/test until production OIDC/JWT is connected

### Tenants, tenancies, and leases

- tenant register with prospect/active/former/blocked lifecycle
- tenant create/list/get/update API
- tenancy model separate from authentication and lease contracts
- primary tenant plus additional occupants
- tenancy create/list/get/update API
- one active tenancy per unit enforced in PostgreSQL
- active tenancy synchronizes unit occupancy
- lease create/list/get/update API
- controlled lease lifecycle transitions
- one active lease per tenancy enforced in PostgreSQL
- rent/deposit amounts stored as integer minor units, never floating point
- manager-facing Tenants and Leasing screens

### Owners and ownership

- owners are business-domain records, separate from application users
- individual and company owner types
- property ownership interests stored in basis points (`10,000 bps = 100%`)
- effective ownership dates retained for historical reporting
- current ownership assignments cannot exceed 100%
- property-row locking serializes concurrent ownership changes
- owner and ownership-interest APIs
- manager-facing Owners workspace

### Rent ledger

- monthly rent obligations derive amount, currency, and due day from the active lease in Go
- one obligation per lease/month
- payments are immutable posted cash records
- allocations are the only mechanism that reduces obligation balances
- obligation balances and open/overdue/paid state are derived from ledger rows
- partial payments are supported
- payment and obligation rows are locked during allocation to prevent concurrent over-allocation
- currency, tenant/tenancy, payment-state, and balance checks are server-authoritative
- manager-facing Rent workspace with obligation generation, payment posting, allocation, receivables, and cash registers

### Maintenance operations

- organization-scoped vendor directory with trade/status controls
- maintenance requests linked to property, optional unit, and optional tenant
- tenant-linked requests require an active occupancy on the selected unit
- work orders support internal work or vendor assignment and scheduling
- request and work-order lifecycle transitions are validated in Go/PostgreSQL transactions
- vendor quotes use integer minor-unit money and require an independent decision permission
- only one approved quote can exist for a work order
- approving a quote atomically assigns its vendor and rejects competing submitted quotes
- completion evidence supports notes today and storage references for future files/photos/invoices
- work orders cannot complete without evidence
- resolved request state follows completed work orders rather than a browser-supplied flag
- quote decisions and evidence capture write actor-aware audit events
- manager-facing Maintenance workspace for intake, dispatch, vendors, quotes, approvals, proof, and lifecycle controls

## Local development

```bash
cp .env.example .env
make db-up
make migrate-all
make seed
```

For an existing database, apply only migrations you have not run:

```bash
make migrate-people
make migrate-finance
make migrate-maintenance
make seed
```

Start the API:

```bash
set -a && source .env && set +a
make api
```

Start the web application separately:

```bash
set -a && source .env && set +a
make web
```

The deterministic development identity is:

```text
Organization: 11111111-1111-1111-1111-111111111111
User:         22222222-2222-2222-2222-222222222222
Role:         admin
```

The deterministic seed includes:

- an active LKR 150,000/month lease
- a September 2026 rent obligation with LKR 50,000 outstanding after allocation
- a 100% property owner relationship
- a resolved plumbing request for the active tenant/unit
- a completed vendor work order
- an approved LKR 18,500 quote
- completion evidence proving the repair and post-work leak test

## API groups

```text
Portfolio
GET/POST       /api/v1/properties
GET/PATCH      /api/v1/properties/{propertyID}
GET/POST       /api/v1/properties/{propertyID}/units
GET/PATCH      /api/v1/units/{unitID}

People
GET/POST       /api/v1/tenants
GET/PATCH      /api/v1/tenants/{tenantID}

Occupancy
GET/POST       /api/v1/tenancies
GET/PATCH      /api/v1/tenancies/{tenancyID}

Contracts
GET/POST       /api/v1/leases
GET/PATCH      /api/v1/leases/{leaseID}

Owners
GET/POST       /api/v1/owners
GET            /api/v1/owners/{ownerID}
GET/POST       /api/v1/ownership-interests

Rent
GET/POST       /api/v1/rent/obligations
GET/POST       /api/v1/rent/payments
POST           /api/v1/rent/allocations

Maintenance
GET/POST       /api/v1/maintenance/vendors
GET/POST       /api/v1/maintenance/requests
PATCH          /api/v1/maintenance/requests/{requestID}/status
GET/POST       /api/v1/maintenance/work-orders
PATCH          /api/v1/maintenance/work-orders/{workOrderID}/status
GET/POST       /api/v1/maintenance/quotes
POST           /api/v1/maintenance/quotes/{quoteID}/decision
GET/POST       /api/v1/maintenance/evidence
```

See `contracts/openapi.yaml` for the request and response contract.

## Development principles

- Keep the backend a **modular monolith** until scaling requirements justify otherwise.
- PostgreSQL is the source of truth.
- Never use floating-point values as persisted money.
- Never trust browser-calculated financial state.
- Derive balances from auditable ledger rows.
- Keep financial approval permissions separate from operational execution permissions.
- Require explicit evidence before declaring operational work complete.
- Keep business rules in Go, not UI components.
- Enforce organization and membership access at the backend boundary.
- Keep tenant/person records separate from portal authentication identities.
- Keep tenancy/occupancy separate from lease contracts.
- Design external authentication and Avara integration behind explicit interfaces.
- Add asynchronous infrastructure only when workflows require it.

## Next domains

1. Production OIDC/JWT identity adapter
2. Documents, object storage, and notifications
3. Resource-scoped owner and tenant portals
4. Rent adjustments, reversals, deposits, owner statements, and reconciliation
5. Maintenance invoice/expense posting and owner approval policies
6. Reporting, audit surfaces, and workflow automation
7. Avara and Blu integration adapters
