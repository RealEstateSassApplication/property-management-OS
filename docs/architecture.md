# Property Management OS: technical architecture

**Reviewed:** main commit 6df541f (2026-10-25). **Scope:** code-grounded architecture, not a certification of a production deployment.

## System overview

Property OS is a multi-tenant, Go modular monolith with a Next.js browser frontend, a separate PostgreSQL-backed notification worker and an optional stdio MCP server. The Go API is the authoritative business and authorization boundary. Database rows are organization-scoped; documents store metadata in PostgreSQL and bytes in S3-compatible storage.

~~~mermaid
flowchart LR
  Browser["Browser: manager, owner or tenant"] --> Web["Next.js App Router / server actions"]
  Web -->|"JSON HTTP + per-request identity"| API["Go HTTP API / domain services"]
  Client["External clients"] -->|"OIDC Bearer + organization"| API
  Partner["Payment provider / adapter"] -->|"Signed HMAC paid event"| API
  Agent["AI client"] -->|"stdio"| MCP["Go MCP server"]
  MCP -->|"Authenticated HTTP API only"| API
  API --> PG[("PostgreSQL: authoritative data")]
  API --> S3[("S3-compatible documents")]
  API --> IDP["OIDC issuer / JWKS"]
  Worker["Go outbox worker"] --> PG
  Worker --> Notify["Webhook notification adapter / Expo push"]
~~~

The deployment configuration is a reference topology, not a complete HA deployment. apps/api/cmd/api starts the HTTP server, apps/api/cmd/worker claims notifications, and apps/api/cmd/mcp exposes only limited API-backed MCP tools. The web lives at apps/web/app.

## Internal layers and module ownership

| Layer | Code | Responsibility |
| --- | --- | --- |
| Transport and RBAC | apps/api/internal/httpapi | Route registration, JWT/organization permission checks, input/output, request IDs |
| Identity | internal/auth, internal/orgadmin, internal/portals | OIDC issuer+subject mapping, organization membership, owner/tenant links |
| Business services | internal/properties, units, tenants, tenancies, leases, owners, rent | Portfolio, occupancy, contracts and receivables |
| Finance | internal/accounting, internal/paymentproviders | Adjustments, deposits, reversals, expenses, owner statement, signed generic provider events |
| Operations | internal/maintenance, internal/inspections | Work orders, quotes, completion evidence and condition reports |
| Supporting | internal/documents, internal/notifications, internal/reporting | Object metadata, durable messaging, derived dashboard/audit |
| AI | internal/agentactions, cmd/mcp | Human-reviewed proposals and constrained API tools |
| Persistence | migrations, domain postgres repositories | Database constraints and organization-scoped queries |
| UI | apps/web/app and apps/web/app/lib | Manager/owner/tenant pages, server actions and per-request API identity |

Keep dependencies client -> handler -> domain service -> repository -> DB. The agent and marketplace must **not** connect directly to authoritative Postgres, bypass permission checks or invent financial state.

## Authentication and tenancy

~~~mermaid
sequenceDiagram
  participant Client as Browser/client
  participant API as HTTP middleware
  participant Identity as OIDC issuer/JWKS
  participant DB as Organization/identity tables
  participant Domain as Domain service
  Client->>API: Bearer token + X-Organization-ID
  API->>Identity: Verify signature/issuer/audience/expiry
  API->>DB: Resolve issuer+subject to internal user
  API->>DB: Verify organization membership and permission
  alt authorized
    API->>Domain: Scoped request with actor identity
    Domain-->>Client: Authorized result
  else denied
    API-->>Client: 401/403
  end
~~~

X-Organization-ID selects a requested scope but **never grants access**. Development-only X-User-ID fallback is not allowed in production. The production web requires a trusted token-forwarding OIDC gateway today; native browser sign-in, refresh/logout and organization switcher still need implementation. Owner and tenant roles use owner_user_links / tenant_user_links for resource-scoped reads, not generalized manager endpoints. See [portals](portals.md) and [deployment](deployment.md).

## Logical data relationships

~~~mermaid
erDiagram
  ORGANIZATIONS ||--o{ PROPERTIES : manages
  ORGANIZATIONS ||--o{ ORGANIZATION_MEMBERSHIPS : authorizes
  PROPERTIES ||--o{ UNITS : contains
  PROPERTIES ||--o{ OWNERSHIP_INTERESTS : has
  OWNERS ||--o{ OWNERSHIP_INTERESTS : holds
  UNITS ||--o{ TENANCIES : occupied_by
  TENANTS ||--o{ TENANCY_TENANTS : joins
  TENANCIES ||--o{ TENANCY_TENANTS : has
  TENANCIES ||--o{ LEASES : governed_by
  LEASES ||--o{ RENT_OBLIGATIONS : bills
  PAYMENTS ||--o{ PAYMENT_ALLOCATIONS : funds
  RENT_OBLIGATIONS ||--o{ PAYMENT_ALLOCATIONS : receives
  MAINTENANCE_REQUESTS ||--o{ WORK_ORDERS : spawns
  WORK_ORDERS ||--o{ MAINTENANCE_QUOTES : estimates
  WORK_ORDERS ||--o{ MAINTENANCE_COMPLETION_EVIDENCE : proves
~~~

This is conceptual. Source-of-truth DDL, keys and constraints are in migrations/000001–000014; additional tables model documents, inspections, accounting, notifications, agent actions, user identity and provider events. The tenancy relationship persists across successive leases. A posted payment and an allocation are distinct objects.

## Business invariants and flows

**Rent:** lease -> one obligation per lease/month -> immutable posted payment -> authorized, row-locked allocations -> outstanding balance derived from ledger. Money is integer minor units with explicit currency. Reversals and adjustments retain evidence; deposits are liabilities, not rent income.

**Maintenance:** request -> work order -> submitted quote -> authorized approval (only one selected quote) -> actual execution -> completion evidence -> completed work order -> resolved request. Quote approval is not proof that work has been done.

**Documents:** API creates authorized metadata and presigned upload -> browser stores bytes -> complete request causes S3 HEAD validation -> metadata available or quarantined -> authorized short-lived download URL. Missing configured bucket means uploads/downloads are unavailable.

**Notifications:** server authors outbox row -> worker claims with FOR UPDATE SKIP LOCKED -> provider delivery -> delivered/retry/dead, with idempotency and bounded retries. Email/SMS/WhatsApp are channel models, not guaranteed configured production channels.

**AI:** MCP read tools inspect authorized context. Narrow write tools create maintenance requests, queue server-authored reminders or propose quote approval. Human user reviews high-risk quote approval in the web UI; the domain service revalidates state. No direct MCP SQL, payment posting or approval authority.

**Onboarding:** current POST /api/v1/onboarding/properties/preview reads limited CSV, validates and reports rows **without persisting anything**. Approved imports are a separate planned workflow.

**Payment events:** generic signed HMAC webhook ingests only normalized paid events and deduplicates by provider/event ID. It does not replace a native payment gateway, refund pipeline or bank settlement reconciliation.

## Isolation and integration rules

Property OS owns the property management operational and financial system of record. The external Avara marketplace (planned integration) owns listings/applications/short-stay booking concerns. Integration must be opt-in, versioned, authenticated, idempotent, auditable, and review conflicts; never dual-write rent or share database access.

Keep the modular monolith until demonstrated scaling, team ownership, security or uptime requirements justify extraction. Prefer horizontal API/worker scaling and safe transaction semantics. Major pending architecture decisions: production browser identity/session, schema migration history, settlement/reconciliation, Avara mapping/events, document retention and privacy policy.

See [engineering handbook](engineering-guide.md), [API contracts](api-and-integrations.md), [feature status](feature-status.md) and [remaining work](implementation-backlog.md).
