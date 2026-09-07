# Property Management OS

Property Management OS is the operational layer for managing rental properties, units, tenants, occupancies, leases, rent, maintenance, documents, vendors, owners, and reporting.

## Architecture

This repository is a monorepo built around a deliberately simple startup architecture:

- **Web:** Next.js + TypeScript
- **Backend:** Go modular monolith
- **Database:** PostgreSQL
- **API contract:** OpenAPI
- **Local development:** Docker Compose

The system is designed as a multi-tenant SaaS from day one. Business data is organization-scoped and authorization is enforced server-side.

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

### Foundation

- Next.js application shell
- Go HTTP API with graceful shutdown and JSON logging
- PostgreSQL baseline schema
- organization-scoped SaaS data model
- OpenAPI contract
- CI for Go and Next.js

### Portfolio

- PostgreSQL connection pooling with pgx
- property list/get/create/update API
- unit list/get/create/update API
- organization scoping on every portfolio query
- property manager portfolio register
- property creation and detail workflows
- occupancy metrics and unit creation

### Identity and authorization

- development identity includes both organization and user context
- every business request verifies `organization_memberships`
- route-level permissions for portfolio, people, and leasing
- admin/manager write access
- accountant/viewer read-only access where appropriate
- maintenance role limited to portfolio visibility
- owner role intentionally denied generalized organization endpoints until owner-resource scoping exists
- development headers disabled outside development/test until a production OIDC/JWT adapter is connected

### Tenants, tenancies, and leases

- tenant register with prospect/active/former/blocked lifecycle
- tenant create/list/get/update API
- tenancy model separate from authentication and lease contracts
- primary tenant plus additional occupant relationship model
- tenancy list/get/create/update API
- one active tenancy per unit enforced in PostgreSQL
- active tenancy synchronizes unit occupancy
- lease contract list/get/create/update API
- controlled lease status transitions
- one active lease per tenancy enforced in PostgreSQL
- rent/deposit amounts stored as integer minor units, never floating point
- manager-facing Tenants and Leasing screens

## Local development

Copy the environment template and start PostgreSQL:

```bash
cp .env.example .env
make db-up
make migrate-all
make seed
```

If you already applied `000001_core` before the people/leasing tranche, run only:

```bash
make migrate-people
make seed
```

Start the API in one terminal:

```bash
set -a && source .env && set +a
make api
```

Start the web application in another terminal:

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

The Next.js server sends those IDs to Go only from server-side requests. Go verifies the membership in PostgreSQL before allowing access. These headers are accepted only when `APP_ENV=development` or `APP_ENV=test`; they are not a production authentication mechanism.

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
```

See `contracts/openapi.yaml` for request and response schemas.

## Development principles

- Keep the backend a **modular monolith** until scaling requirements justify otherwise.
- PostgreSQL is the source of truth.
- Never use floating-point values as persisted money.
- Keep business rules in Go, not UI components.
- Enforce organization and membership access at the backend boundary.
- Keep tenant/person records separate from portal authentication identities.
- Keep tenancy/occupancy separate from lease contracts.
- Design external authentication and Avara integration behind explicit interfaces.
- Add asynchronous infrastructure only when the workflow requires it.

## Next domains

1. Production OIDC/JWT identity adapter
2. Owners and property ownership relationships
3. Rent obligations, ledger, payment allocation, and arrears
4. Maintenance requests, work orders, vendors, and approvals
5. Documents and notifications
6. Owner and tenant portals with resource-scoped authorization
7. Reporting, audit surfaces, and workflow automation
