# Property Management OS

Property Management OS is a multi-tenant operational platform for rental property portfolios: properties and units, owners, tenants and occupancies, leases, rent, maintenance, documents, notifications, and agent-assisted operations.

## Use Cases

Property Management OS is designed for **residential property-management companies, independent property managers, and teams managing rental portfolios on behalf of multiple owners**. It brings portfolio records, leasing, rent collection, maintenance, documents, and operational oversight into one organization-scoped system.

### Who uses it and why

| User / scenario | What they can do with Property OS | Implementation status |
| --- | --- | --- |
| **Property manager — manage a portfolio** | Register properties and units, monitor occupancy, organize owners, and view operational portfolio metrics. | **Implemented foundation** |
| **Leasing team — move a tenant into a unit** | Maintain tenant records, assign an occupancy/tenancy, manage lease terms and track the lease lifecycle without conflating tenants with login identities. | **Implemented foundation** |
| **Accountant — track rent and cash** | Generate lease-based monthly obligations, post payments, allocate partial/full receipts, inspect arrears, record adjustments/reversals and handle security deposits. | **Implemented foundation**; settlement reconciliation requires further work |
| **Property owner — review managed assets** | Access an owner-scoped portal summary with linked properties, occupancy, outstanding receivables and maintenance counts; managers can generate accounting owner statements. | **Implemented foundation**; evidence-backed **Owner Proofbook** and export are planned |
| **Tenant — view tenancy and request maintenance** | See linked occupancy, lease/rent information and maintenance items, and submit a maintenance request tied to an authorized tenancy. | **Implemented resource-scoped portal foundation** |
| **Maintenance coordinator — resolve property issues** | Intake and triage requests, assign work orders, handle vendor quotes and approvals, attach completion evidence and close completed work. | **Implemented foundation** |
| **Field/inspection team — document condition** | Record move-in, move-out and periodic inspections, condition items, evidence references, completion and acknowledgement. | **Implemented foundation**; mobile/offline field UX needs independent validation |
| **Operations staff — store files and send reminders** | Attach lease/property/maintenance documents, request controlled downloads, queue ledger-derived rent reminders and inspect delivery status. | **Implemented foundations**; real S3 storage and delivery providers must be configured and validated |
| **Platform administrator — control access and monitor operations** | Manage organization settings, members, roles, owner/tenant links, dashboards and audit-event views. | **Implemented foundation**; complete self-service signup/login and SaaS billing are unfinished |
| **AI-assisted operator — surface issues and propose action** | Use MCP tools to review arrears, expiring leases and open maintenance; queue controlled reminders or propose quote approvals for human review. | **Implemented constrained tools**; agents cannot directly approve high-risk spending or post money |

**Example end-to-end workflow:** A manager registers a property and unit → creates a tenant and active tenancy → activates a lease → generates the monthly rent obligation → records a payment and allocates it → reviews outstanding rent and owner reporting. Another workflow starts with a tenant maintenance request → manager work order and vendor quote → authorized human approval → recorded completion evidence.

### Planned or incomplete use cases

- **Switching from spreadsheets or another property-management system:** The current onboarding flow **only validates and previews property CSV files**; approved, idempotent imports, collision review and reconciliation are planned ([Issue #20](https://github.com/RealEstateSassApplication/property-management-OS/issues/20)).
- **Sending owners a complete evidence-backed report:** The **Owner Proofbook** would connect rent, adjustments, maintenance approvals/completion and inspection evidence into scoped, exportable reports ([Issue #19](https://github.com/RealEstateSassApplication/property-management-OS/issues/19)).
- **Receiving applicants from Avara Marketplace:** A consented, reviewed application handoff and safe event integration are planned; Avara booking data must not become a lease or a duplicate financial ledger ([Issue #21](https://github.com/RealEstateSassApplication/property-management-OS/issues/21)).
- **Real-world customer onboarding and payment processing:** Native web login/session lifecycle, paid subscription billing, provider-specific payment integration, bank settlement matching, production notifications and deployment sign-off still need development and validation.

**Scope note:** “Implemented foundation” means the repository contains the core domain APIs, persistence and/or UI for that scenario—not that the entire workflow has passed production security, financial reconciliation, provider testing or customer acceptance. See [feature status](docs/feature-status.md) and [remaining implementation work](docs/implementation-backlog.md).

## Product strategy, Avara integration and roadmap

Read [the Avara + Property OS competitive strategy and phased roadmap](docs/product-strategy-and-roadmap.md). It identifies the initial residential property-manager customer, the planned Owner Proofbook differentiation, release criteria, and the intended separation between Property OS financial records and the Avara marketplace.

The first onboarding milestone is **read-only property CSV preflight**: open `/onboarding` in the web dashboard, use [the sample CSV](docs/samples/properties-onboarding.csv), and inspect the validation report. The authenticated `POST /api/v1/onboarding/properties/preview` API requires `portfolio:manage`, accepts `text/csv` (512 KiB / 500 rows maximum), and **does not import, store, create or update any data**. See [the API contract](contracts/onboarding.openapi.yaml). A reviewed and reconciled import workflow is planned, not yet implemented.

**Production caveat:** Feature implementation and CI success do not substitute for security review, real-user permissions, financial reconciliation, provider staging tests, documented restore exercises and pilot acceptance. Follow [production readiness](docs/production-readiness.md) before using real customer data.

## Engineering documentation and remaining work

Start here for a code-grounded engineering handoff and product completion plan (audited against main 6df541f on 2026-10-25):

- [Engineering handbook](docs/engineering-guide.md) — repository map, local environment, API identity, contributor workflow, debugging.
- [Architecture and data-flow diagrams](docs/architecture.md) — processes, trust boundaries, logical data model and key workflows.
- [Feature status inventory](docs/feature-status.md) — implemented vs partial vs planned vs not verified.
- [Prioritized remaining work](docs/implementation-backlog.md) — P0 launch requirements, P1 pilot features, P2 later expansions.
- [API and integration reference](docs/api-and-integrations.md) — routes, contract files and provider ownership.
- [Deployment and incident runbook](docs/operations-runbook.md) — releases, monitoring, backups and recovery.
- [QA and release checklist](docs/qa-and-release-checklist.md) — security, tenancy, financial and E2E acceptance criteria.

**Product caveat:** substantial foundations are implemented, but real-world production readiness remains unverified. Native browser authentication/session UX, approved imports, evidence-backed Owner Proofbook and consented Avara integration are incomplete.

## Architecture

Property Management OS uses a **multi-tenant modular monolith**. The Next.js application is the user-facing interface; the Go HTTP API owns authorization, domain rules and all authoritative writes. PostgreSQL stores operational and financial records, while an S3-compatible bucket stores document bytes. A dedicated Go worker delivers notifications asynchronously. The optional MCP server exposes narrowly scoped agent tools **through the same authenticated API**, never directly through database access.

### System architecture

~~~mermaid
flowchart TB
    subgraph Clients["Clients and integrations"]
        Browser["Manager / Owner / Tenant browser"]
        Agent["AI agent / IDE"]
        Provider["Payment provider adapter"]
    end

    subgraph Application["Property Management OS"]
        Gateway["Trusted identity gateway / OIDC session handling"]
        Web["Next.js web app"]
        API["Go HTTP API / modular monolith"]
        MCP["Go MCP server (stdio)"]
        Worker["Go notification worker"]
    end

    subgraph Data["Persistence and external services"]
        DB[("PostgreSQL: domain data, ledgers, outbox")]
        Store[("S3-compatible object storage")]
        OIDC["OIDC issuer / JWKS"]
        Notify["Notification provider / Expo Push"]
    end

    Browser -->|"HTTPS"| Gateway
    Gateway --> Web
    Gateway -.->|"OIDC integration"| OIDC
    Web -->|"Per-request Bearer token + organization scope"| API
    Agent -->|"MCP stdio"| MCP
    MCP -->|"Authenticated HTTP API"| API
    Provider -->|"Signed payment-paid webhook"| API
    API -->|"Verify token via OIDC"| OIDC
    API <--> DB
    API -->|"Presigned upload / download flow"| Store
    Worker -->|"Claim notification outbox"| DB
    Worker -->|"Send / retry"| Notify
~~~

> **Authentication note:** The API verifies OIDC/JWT tokens and organization membership today. A trusted authenticated gateway can supply browser-request identity, but built-in browser login, refresh/logout and organization switching are still productization work. Development \`X-User-ID\` fallback must never be exposed publicly. The payment endpoint is a generic signed adapter, not a complete external checkout/reconciliation integration.

### Backend modules

The single Go API in \`apps/api/cmd/api\` composes domain services from \`apps/api/internal/\`. They are organized around business capabilities rather than deployed as individual microservices.

| Domain | Primary responsibility | Source |
| --- | --- | --- |
| Identity and administration | OIDC mapping, organization membership, RBAC, settings, portal links | \`auth/\`, \`orgadmin/\`, \`portals/\` |
| Portfolio | Properties, units, occupancy and ownership interests | \`properties/\`, \`units/\`, \`owners/\` |
| Leasing | Tenants, tenancies and lease lifecycle | \`tenants/\`, \`tenancies/\`, \`leases/\` |
| Finance | Rent obligations, posted payments, allocations, deposits, expenses and owner statements | \`rent/\`, \`accounting/\`, \`paymentproviders/\` |
| Operations | Maintenance requests, vendors, quotes, work orders, evidence and inspections | \`maintenance/\`, \`inspections/\` |
| Platform services | Document storage, notification outbox, reporting and audit | \`documents/\`, \`notifications/\`, \`reporting/\` |
| Agent workflows | Limited MCP tools, proposals and human-reviewed approvals | \`agentactions/\`, \`cmd/mcp/\` |

### Core business relationships

~~~mermaid
flowchart LR
    Org["Organization"] --> Property["Property"]
    Property --> Unit["Unit"]
    Property --> OwnerInterest["Ownership interests / owners"]
    Unit --> Tenancy["Tenancy / occupants"]
    Tenancy --> Lease["Lease"]
    Lease --> Obligation["Monthly rent obligation"]
    Payment["Posted payment"] --> Allocation["Payment allocation"]
    Allocation --> Obligation
    Property --> Request["Maintenance request"]
    Request --> Work["Work order"]
    Work --> Quote["Vendor quote / human approval"]
    Work --> Evidence["Completion evidence"]
~~~

This is a **logical model**, not the exact SQL schema. Read [domain-model.md](docs/domain-model.md) and the numbered \`migrations/\` files for constraints, table relationships and lifecycle rules.

### How the key workflows operate

1. **Authentication and isolation:** Requests carry the current user's Bearer token and an organization selector. Go verifies OIDC issuer/subject, resolves the internal user and checks organization membership plus route permissions. Owner and tenant portals additionally restrict access through explicit resource links.
2. **Rent and accounting:** The backend derives obligations from leases and stores payments separately from allocations. Outstanding balances come from auditable records; financial amounts use integer minor units and preserve reversals.
3. **Maintenance and inspections:** Requests produce work orders, quotes require appropriate approvals, and a work order cannot be completed without evidence. Document files live in object storage with controlled metadata and authorized signed URLs.
4. **Notifications:** API services enqueue a durable PostgreSQL outbox record. The worker claims jobs, sends through configured delivery providers and marks deliveries, retries or terminal failures.
5. **AI-agent operations:** MCP calls the ordinary authenticated API. Low-risk actions remain constrained; quote-approval proposals require review by an authorized human before execution.
6. **External integrations:** A generic signed payment webhook exists; the optional Avara marketplace handoff remains planned. No external integration gets direct access to Property OS financial tables.

### Detailed diagrams and engineering documentation

- [Architecture and trust boundaries](docs/architecture.md)
- [Domain model and lifecycle rules](docs/domain-model.md)
- [Engineering handbook and local setup](docs/engineering-guide.md)
- [API and external integrations](docs/api-and-integrations.md)
- [Implemented features vs. unfinished work](docs/feature-status.md)
- [Prioritized implementation backlog](docs/implementation-backlog.md)

Core rules:

- **Web:** Next.js + TypeScript
- **Backend:** Go modular monolith
- **Database:** PostgreSQL
- **Authentication:** standards-compliant OIDC / Bearer JWT
- **Authorization:** PostgreSQL organization memberships + RBAC
- **Documents:** S3-compatible storage with PostgreSQL metadata
- **Async delivery:** PostgreSQL transactional outbox + Go worker
- **Agent integration:** official Go MCP SDK over the authenticated HTTP API
- **Contracts:** OpenAPI
- **Local development:** Docker Compose

Business data is organization-scoped and authorization is enforced server-side. The MCP process has no database connection and does not provide an alternate authorization path.

## Repository layout

```text
apps/
  web/                 Next.js property-manager experience
  api/
    cmd/api/           HTTP API
    cmd/worker/        notification delivery worker
    cmd/mcp/           stdio MCP server for agentic operations
contracts/
  openapi.yaml
  documents.openapi.yaml
  notifications.openapi.yaml
migrations/
scripts/
docs/
```

## Implemented

### Foundation, identity, and authorization

- Go HTTP API with structured logging and graceful shutdown
- PostgreSQL connection pooling with pgx
- organization-scoped multi-tenancy
- production OIDC discovery/JWKS/JWT verification
- external `issuer + subject` mapped to stable internal users
- PostgreSQL organization membership and role permissions
- development/test `X-User-ID` fallback only when a Bearer token is not supplied
- production refuses to start without complete OIDC configuration

Production authorization chain:

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

`X-Organization-ID` selects requested tenant scope; it never grants access by itself.

Current role intent:

- `admin`, `manager`: manage all implemented organization modules
- `accountant`: read operational/financial data, manage rent, approve maintenance costs, view notification delivery, and queue server-authored rent reminders
- `viewer`: read-only operational access, excluding generalized sensitive documents/notification delivery
- `maintenance`: portfolio visibility and maintenance execution, but no vendor master-data management or cost approval
- `owner`: no generalized organization access until resource-scoped owner portal authorization is implemented

### Portfolio

- properties and units
- create/list/get/update APIs
- occupancy state
- portfolio register, property details, unit management and metrics

### Tenants, tenancies, and leases

- tenant lifecycle
- tenancy separated from login identity and lease contracts
- primary tenant plus additional occupants
- one active tenancy per unit enforced in PostgreSQL
- active tenancy synchronizes unit occupancy
- controlled lease lifecycle
- one active lease per tenancy
- rent/deposit money stored as integer minor units
- manager-facing Tenants and Leasing workspaces

### Owners and ownership

- owner domain records separate from application users
- individual/company owners
- ownership interests in basis points (`10,000 bps = 100%`)
- effective ownership dates
- property-row locking prevents concurrent current interests exceeding 100%
- manager-facing Owners workspace

### Rent ledger

- monthly obligations derive amount, currency and due day from the lease in Go
- one obligation per lease/month
- immutable posted payments
- allocations are the only mechanism reducing obligation balances
- open/overdue/paid state and balances derived from ledger rows
- partial payments
- payment/obligation row locking during allocation
- server-authoritative currency, tenancy, payment-state and balance validation
- manager-facing Rent receivables and cash registers

### Maintenance operations

- organization-scoped vendor directory
- property/unit/tenant-linked requests
- active-occupancy validation for tenant-linked requests
- work orders, assignment and scheduling
- quote approval separated from execution permission
- one approved quote per work order
- approved quote atomically assigns its vendor and rejects competing submitted quotes
- completion evidence required before work-order completion
- actor-aware audit events
- manager-facing Maintenance workspace

### Documents and object storage

- organization/resource-scoped document metadata in PostgreSQL
- S3-compatible object storage for file bytes
- backend-generated storage keys
- presigned browser-direct uploads
- 25 MB server-authoritative limit and MIME allowlist
- HEAD verification before `pending → available`
- mismatch quarantine
- short-lived presigned downloads
- storage deletion plus metadata tombstone
- attachments for properties, units, tenants, leases, owners, rent payments, maintenance requests, work orders and vendors
- generalized document access remains admin/manager-only until resource-scoped portals are implemented
- manager-facing Documents workspace

### Notification outbox and worker

- durable `notification_outbox` records
- email/SMS/WhatsApp/webhook channel model
- `pending → processing → delivered`, with `retry` and `dead` failure states
- idempotency keys
- concurrent worker claiming with `FOR UPDATE SKIP LOCKED`
- stale processing-lock recovery
- capped exponential retries
- development log provider
- production webhook delivery adapter boundary
- manager-facing Notifications delivery register
- server-authored rent reminders based on the real ledger state

Rent-reminder callers supply only obligation ID, channel and recipient. The API derives tenant, property, unit, current outstanding balance and due date and rejects paid/void obligations. Same-day reminders are idempotent.

### Agentic MCP server

Property OS includes a stdio MCP server at `apps/api/cmd/mcp`. It uses the official `modelcontextprotocol/go-sdk` and calls the authenticated Property OS API rather than importing repositories.

Read tools:

- `portfolio_snapshot`
- `list_overdue_rent`
- `list_expiring_leases`
- `list_open_maintenance`

Controlled mutation tools:

- `queue_rent_reminder`
- `create_maintenance_request`

The first MCP tranche deliberately excludes generic SQL, arbitrary financial writes, payment posting, quote approval and blanket document access. See `docs/mcp.md`.

## Local development

```bash
cp .env.example .env
make db-up
make migrate-all
make seed
```

For an existing database, apply only migrations you have not run, including:

```bash
make migrate-documents
make migrate-notifications
```

Run the API, web app and async worker in separate terminals:

```bash
set -a && source .env && set +a
make api
```

```bash
set -a && source .env && set +a
make web
```

```bash
set -a && source .env && set +a
make worker
```

Run the MCP server for a local MCP client:

```bash
set -a && source .env && set +a
make mcp
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

Production API authentication uses:

```http
Authorization: Bearer <verified OIDC JWT>
X-Organization-ID: <organization UUID>
```

For a production MCP process, configure a short-lived `PROPERTY_OS_ACCESS_TOKEN` for the intended user/service identity rather than `PROPERTY_OS_USER_ID`. Never commit access tokens into the repository or MCP configuration files.

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

Documents
GET            /api/v1/documents
POST           /api/v1/documents/uploads
POST           /api/v1/documents/{documentID}/complete
GET            /api/v1/documents/{documentID}/download
DELETE         /api/v1/documents/{documentID}

Notifications
GET/POST       /api/v1/notifications
POST           /api/v1/notifications/rent-reminders
```

See `contracts/openapi.yaml`, `contracts/documents.openapi.yaml`, and `contracts/notifications.openapi.yaml`.

## Development principles

- Keep the backend a **modular monolith** until scaling requirements justify otherwise.
- PostgreSQL is the source of truth for domain and outbox state.
- Separate authentication from authorization.
- Keep external identities separate from stable internal user IDs.
- Never authorize by email alone.
- Never use floating-point values as persisted money.
- Never trust browser- or agent-calculated financial state.
- Derive balances from auditable ledger rows.
- Keep financial approval separate from operational execution.
- Require explicit evidence before declaring operational work complete.
- Keep tenant/person records separate from portal authentication identities.
- Keep tenancy/occupancy separate from lease contracts.
- Send external notifications asynchronously from durable outbox state.
- Agents call authenticated domain APIs; they do not receive raw database access.

## Remaining product milestones

Resource-scoped portals, accounting adjustments/reversals, deposits/expenses, reporting and human-reviewed agent proposals already have code implementations. The remaining work is **not** to rewrite those foundations: complete the migration and identity lifecycle, prove data/finance invariants in staging, deliver real provider reconciliation, build approved CSV import, deliver Owner Proofbook and integrate Avara safely. Follow [feature status](docs/feature-status.md) and the [prioritized backlog](docs/implementation-backlog.md).
