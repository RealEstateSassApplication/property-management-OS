# Architecture

## Goal

Property Management OS is a multi-tenant property operations platform. It owns the operational lifecycle after a property enters management: portfolio structure, tenancy, leases, rent, maintenance, documents, vendors, reporting, and automation.

Avara and other marketplace/transaction products integrate through explicit APIs and events. They must not read or write this application's database directly.

## System shape

```text
Browser / future mobile clients / partner systems
                    |
                    v
              Next.js Web
                    |
              HTTPS / JSON
                    |
                    v
               Go API
       modular monolith boundary
                    |
          +---------+---------+
          |                   |
          v                   v
     PostgreSQL          Object storage
    source of truth       documents/media
```

Background work initially runs from a separate Go worker entrypoint in the same codebase. A dedicated queue or broker should only be introduced when delivery, scale, or retry requirements make it necessary.

## Backend module boundaries

The Go application is organized by business capability rather than technical layers alone:

- auth
- organizations
- properties
- units
- owners
- tenants
- tenancies
- leases
- rent
- payments
- maintenance
- vendors
- documents
- notifications
- reporting
- audit

Modules may expose application services and interfaces to other modules. Database tables are not an excuse to bypass module rules.

## Multi-tenancy

`organization_id` is the primary SaaS tenant boundary.

Rules:

1. Every organization-owned resource must be scoped server-side.
2. Client-provided organization IDs are never sufficient authorization.
3. Membership and role checks happen before business operations.
4. Queries for tenant-owned data include the organization boundary.
5. Audit events record security-sensitive and financially relevant changes.

## API

REST is the initial external interface and is described by `contracts/openapi.yaml`.

Conventions:

- Version business endpoints under `/api/v1`.
- Use server-authoritative calculations for money and state transitions.
- Prefer idempotency keys for payment/integration write operations.
- Return structured problem responses with stable error codes.
- Propagate request IDs into logs and audit events.

## Data

PostgreSQL is the system of record. Financial history should be modeled as append-friendly ledgers rather than mutable totals. Derived dashboard values should be calculated from authoritative records or maintained projections.

Object/blob storage will hold documents and media. The database stores metadata, ownership, integrity, and lifecycle state.

## Authentication and authorization

Authentication may be delegated to a managed identity provider, but authorization remains inside Property Management OS. The Go backend maps the authenticated identity to a local user and organization memberships.

Initial roles:

- admin
- manager
- owner
- accountant
- maintenance
- viewer

Tenant-facing access should be modeled separately when the tenant portal is implemented rather than granting broad organization membership.

## Avara integration

Integration is one-way or event/API based at explicit boundaries. Example property import flow:

```text
Avara -> POST integration command -> Property OS validates -> creates/links property -> emits audit/integration result
```

The `external_avara_property_id` field supports identity correlation without coupling storage schemas.

## Scaling strategy

Do not split into microservices by default. First scale the modular monolith horizontally and isolate expensive background work. Extract a service only when a module has a demonstrated independent scaling, security, ownership, or availability requirement.

## Observability

The target baseline is structured JSON logging, request IDs, metrics, traces, and OpenTelemetry-compatible instrumentation. Health and readiness endpoints should remain lightweight and suitable for container orchestration.
