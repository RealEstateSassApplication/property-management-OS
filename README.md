# Property Management OS

Property Management OS is the operational layer for managing rental properties, units, owners, tenants, occupancies, leases, rent, maintenance, documents, vendors, and reporting.

## Architecture

This repository is a monorepo built around a deliberately simple startup architecture:

- **Web:** Next.js + TypeScript
- **Backend:** Go modular monolith
- **Database:** PostgreSQL
- **Authentication:** standards-compliant OIDC / Bearer JWT
- **Authorization:** PostgreSQL organization memberships + RBAC
- **API contract:** OpenAPI
- **Local development:** Docker Compose

The system is multi-tenant from day one. Business data is organization-scoped and all authorization is enforced server-side.

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
- CI for Go, PostgreSQL, and Next.js

### Production identity and authorization

Production authentication is provider-neutral OIDC. The API performs issuer discovery, verifies JWT signatures against the provider JWKS, validates issuer/audience/token lifetime, and then maps the verified external identity to a stable internal Property OS user.

```text
Bearer JWT
   ↓
OIDC issuer / JWKS / audience verification
   ↓
verified issuer + subject
   ↓
user_identities
   ↓
internal users.id
   ↓
organization_memberships
   ↓
route permission
```

Important rules:

- an OIDC `sub` is never used as a Property OS user UUID
- `issuer + subject` uniquely identify a linked external identity
- email is retained as last-seen metadata but is never used to grant access
- a valid token whose identity is not linked returns `403 Forbidden`
- an invalid, expired, or unverifiable Bearer token returns `401 Unauthorized`
- `X-Organization-ID` selects the organization the caller wants to operate in; it does **not** grant membership
- PostgreSQL `organization_memberships` remains the authorization authority
- production startup fails if `OIDC_ISSUER_URL` and `OIDC_AUDIENCE` are not configured together
- development/test may still use `X-User-ID` when no Bearer token is supplied
- if a Bearer token is supplied in development, the real OIDC path takes precedence

Current roles and permissions include:

- `admin`, `manager`: manage all implemented organization modules
- `accountant`: read operational data, manage rent, and approve maintenance costs
- `viewer`: read-only operational access
- `maintenance`: portfolio visibility and maintenance execution, but not vendor master-data management or cost approval
- `owner`: no generalized organization access until resource-scoped owner portal authorization is implemented

### Tenants, tenancies, and leases

- tenant register with prospect/active/former/blocked lifecycle
- tenancy model separate from authentication and lease contracts
- primary tenant plus additional occupants
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
- ownership interests stored in basis points (`10,000 bps = 100%`)
- effective ownership dates retained for historical reporting
- current ownership assignments cannot exceed 100%
- property-row locking serializes concurrent ownership changes
- manager-facing Owners workspace

### Rent ledger

- monthly obligations derive amount, currency, and due day from the active lease in Go
- one obligation per lease/month
- immutable posted payment records
- allocations are the only mechanism that reduces obligation balances
- balances and open/overdue/paid state are derived from ledger rows
- partial payments supported
- payment and obligation rows locked during allocation
- currency, tenancy, payment-state, and balance checks are server-authoritative
- manager-facing Rent workspace with receivables and cash registers

### Maintenance operations

- organization-scoped vendor directory
- property/unit/tenant-linked maintenance requests
- tenant-linked requests require an active occupancy on the selected unit
- work orders support internal or vendor assignment and scheduling
- quote money is integer minor-unit data
- cost approval is separate from maintenance execution permission
- one approved quote per work order
- approving a quote atomically assigns its vendor and rejects competing submitted quotes
- work orders cannot complete without completion evidence
- request resolution follows completed work orders
- quote decisions and evidence capture retain actor-aware audit events
- manager-facing Maintenance workspace for intake, dispatch, quotes, approvals, proof, and lifecycle controls

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
make migrate-identity
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

Development identity:

```text
Organization: 11111111-1111-1111-1111-111111111111
User:         22222222-2222-2222-2222-222222222222
Role:         admin
```

The deterministic seed also contains a test external identity mapping:

```text
Issuer:  https://idp.example.invalid
Subject: property-os-dev-user
User:    22222222-2222-2222-2222-222222222222
```

For production set:

```bash
APP_ENV=production
OIDC_ISSUER_URL=https://your-idp.example.com/
OIDC_AUDIENCE=your-property-os-api-audience
```

The configured provider may be Auth0, Keycloak, Okta, Microsoft Entra ID, or another OIDC-compliant issuer.

## API authentication

Production business requests use:

```http
Authorization: Bearer <verified OIDC JWT>
X-Organization-ID: <organization UUID>
```

The organization header is a requested tenant scope. After authentication, the backend verifies that the resolved internal user is an active member with the permission required by the route.

Development/test can instead use:

```http
X-Organization-ID: <organization UUID>
X-User-ID: <internal user UUID>
```

Those development headers are not accepted as production authentication.

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

See `contracts/openapi.yaml` for the request/response contract.

## Development principles

- Keep the backend a **modular monolith** until scaling requirements justify otherwise.
- PostgreSQL is the source of truth.
- Separate authentication from authorization.
- Keep external identities separate from stable internal user IDs.
- Never authorize by email alone.
- Never use floating-point values as persisted money.
- Never trust browser-calculated financial state.
- Derive balances from auditable ledger rows.
- Keep financial approval separate from operational execution.
- Require explicit evidence before declaring operational work complete.
- Keep tenant/person records separate from portal authentication identities.
- Keep tenancy/occupancy separate from lease contracts.
- Add asynchronous infrastructure only when workflows require it.

## Next domains

1. Documents, object storage, and notifications
2. Resource-scoped owner and tenant portals
3. Rent adjustments, reversals, deposits, owner statements, and reconciliation
4. Maintenance invoice/expense posting and owner approval policies
5. Reporting, audit surfaces, and workflow automation
6. Avara and Blu integration adapters
