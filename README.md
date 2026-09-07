# Property Management OS

Property Management OS is the operational layer for managing rental properties, units, owners, tenants, leases, rent, maintenance, documents, vendors, and reporting.

## Architecture

This repository is a monorepo built around a deliberately simple startup architecture:

- **Web:** Next.js + TypeScript
- **Backend:** Go modular monolith
- **Database:** PostgreSQL
- **API contract:** OpenAPI
- **Local development:** Docker Compose

The system is designed as a multi-tenant SaaS from day one. Business data is scoped to an organization and authorization must be enforced server-side.

## Repository layout

```text
apps/
  web/                 Next.js property manager, owner, and tenant experiences
  api/                 Go API and background worker entrypoints
contracts/
  openapi.yaml         Source API contract
migrations/            PostgreSQL migrations
scripts/               Local development helpers and seed data
infra/                 Deployment and infrastructure assets
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

### Portfolio vertical slice

- PostgreSQL connection pooling with pgx
- property list/get/create/update API
- property unit list/get/create/update API
- organization scoping on every portfolio query
- development-only organization identity header (disabled outside development/test)
- property/unit validation and unit tests
- property manager portfolio register
- property creation flow
- property detail and occupancy metrics
- unit register and unit creation flow
- deterministic local seed organization/property/units

## Local development

Copy the environment template and start PostgreSQL:

```bash
cp .env.example .env
make db-up
make migrate-up
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

The development seed uses organization ID:

```text
11111111-1111-1111-1111-111111111111
```

The web server sends that ID to the Go API server-side. The API accepts this development identity only when `APP_ENV=development` or `APP_ENV=test`. Production endpoints intentionally remain unavailable until authenticated identity/RBAC is connected.

## Portfolio API

```text
GET    /api/v1/properties
POST   /api/v1/properties
GET    /api/v1/properties/{propertyID}
PATCH  /api/v1/properties/{propertyID}
GET    /api/v1/properties/{propertyID}/units
POST   /api/v1/properties/{propertyID}/units
GET    /api/v1/units/{unitID}
PATCH  /api/v1/units/{unitID}
```

See `contracts/openapi.yaml` for request/response schemas.

## Development principles

- Keep the backend a **modular monolith** until scaling requirements justify otherwise.
- PostgreSQL is the source of truth.
- Never trust client-calculated financial values.
- Keep business rules in the Go backend, not in UI components.
- Enforce organization/ownership access at the backend boundary.
- Design all external integrations behind explicit interfaces.
- Prefer OpenAPI-generated clients/types between Go and Next.js.
- Add asynchronous infrastructure only when the workflow requires it.

## Next domains

1. Authentication provider + organization memberships/RBAC
2. Owners and ownership relationships
3. Tenants and tenancies
4. Leases, renewals, deposits, and termination
5. Rent obligations, ledger, payments, and arrears
6. Maintenance requests, work orders, vendors, and approvals
7. Documents, notifications, reporting, and owner portal
