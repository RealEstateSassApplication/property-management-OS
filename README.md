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
infra/                  Deployment and infrastructure assets
docs/                   Architecture and domain documentation
```

## Initial domains

1. Organizations and RBAC
2. Properties and units
3. Owners
4. Tenants
5. Leases and tenancies
6. Rent ledger and payments
7. Maintenance and vendors
8. Documents
9. Notifications
10. Reporting and audit logs

## Development principles

- Keep the backend a **modular monolith** until scaling requirements justify otherwise.
- PostgreSQL is the source of truth.
- Never trust client-calculated financial values.
- Keep business rules in the Go backend, not in UI components.
- Enforce organization/ownership access at the backend boundary.
- Design all external integrations behind explicit interfaces.
- Prefer OpenAPI-generated clients/types between Go and Next.js.
- Add asynchronous infrastructure only when the workflow requires it.

## Status

Bootstrap in progress. The first tranche establishes the repository structure, local development environment, API contract, health endpoints, database baseline, and application shells before domain implementation begins.
