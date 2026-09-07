# Property Management OS

Property Management OS is a multi-tenant operational platform for rental property portfolios: properties and units, owners, tenants and occupancies, leases, rent, maintenance, documents, notifications, and agent-assisted operations.

## Architecture

```text
Browser / manager UI
        |
        v
Next.js web
        |
        v
Go API modular monolith ---------------------------+
        |                                           |
        +--> PostgreSQL                             |
        +--> S3-compatible object storage           |
        +--> notification_outbox                    |
                                                    |
Go notification worker <----------------------------+
        |
        +--> delivery provider (log dev / webhook production)

AI agent / IDE
        |
        v
Property OS MCP (stdio)
        |
        v
Authenticated Property OS HTTP API
```

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

## Next domains

1. Resource-scoped owner and tenant portals
2. Scheduled operational automation and approval policies for agent actions
3. Rent adjustments, reversals, deposits, owner statements and reconciliation
4. Maintenance invoice/expense posting and owner approval policies
5. Reporting, audit surfaces and operational analytics
6. Avara and Blu integration adapters
