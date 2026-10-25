# Property OS engineering and developer handbook

**Baseline:** main 6df541f, reviewed 2026-10-25. This guide consolidates setup, code ownership, workflows and documentation. The production environment itself was not exercised.

## Product and repository

The platform is a multi-organization, residential property operations SaaS aimed initially at property managers who handle properties, owners, tenants, leasing, rent, expenses, maintenance and evidence. Separate Avara marketplace integration is optional and unfinished.

~~~text
apps/api/cmd/api       Go HTTP API
apps/api/cmd/worker    Go notification background worker
apps/api/cmd/mcp       stdio MCP server
apps/api/internal/     domain models, services, repositories, HTTP and tests
apps/web/app/          Next.js 16 + TypeScript + React 19 App Router pages
apps/web/app/lib/      request-scoped API identity and domain clients
contracts/             domain OpenAPI documents
migrations/            ordered PostgreSQL SQL up/down files
scripts/               local bootstrap, seed and backup/restore helpers
.github/workflows/     CI
docs/                  domain and operational guides
~~~

## Fresh local setup

Requirements: Docker with Compose, make, Node.js 24 with npm, and the Go version in apps/api/go.mod.

~~~bash
cp .env.example .env
make dev-setup
make dev
~~~

Web: http://localhost:3000; API: http://localhost:8080; /healthz and /readyz. The seeded development organization is 11111111-1111-1111-1111-111111111111; seeded admin user is 22222222-2222-2222-2222-222222222222. Neither is a production identity.

Manual boot for a **fresh** database:

~~~bash
make db-up
make migrate-all
make seed
# Load .env into the shell first, then in separate terminals:
make api
make worker
make web
~~~

**Existing DB warning:** scripts/dev-setup.sh detects the organizations table and skips all migrations. It is **not** a schema migration tracker. Never rerun the entire migration chain on an existing DB: first determine which numbered migrations are missing and back up data. In a database missing only push-device migration 000014, use make migrate-push-devices (after verifying it was not already applied). Adopt a durable versioned migration runner before automatic production upgrades.

Local document file operations return 503 if S3 storage is unconfigured. Notification provider defaults to console logging only in development/test; production requires a configured delivery adapter. Blank PAYMENT_WEBHOOK_SECRET disables the generic payment endpoint. The default make dev launches API, worker and web but **not** MCP; run make mcp separately for an MCP client.

## Routine checks

~~~bash
make test-api
make build-api
npm install
npm run lint:web
npm run test:web
npm run build:web
POSTGRES_PASSWORD=example docker compose -f docker-compose.production.yml config
~~~

The last command only validates production-style Compose syntax with a placeholder. GitHub Actions exercises migrations and seed invariants; review the CI runs of the exact release SHA. Do not claim integration/provider/identity readiness based on unit tests.

## Local authenticated request

~~~bash
curl -fsS http://localhost:8080/api/v1/properties \
  -H 'X-Organization-ID: 11111111-1111-1111-1111-111111111111' \
  -H 'X-User-ID: 22222222-2222-2222-2222-222222222222'
~~~

X-User-ID is development/test only. In production, the API verifies an OIDC Bearer token (issuer, audience, signature, expiry and identity mapping) and membership with the requested organization. The web needs actual per-request identity from an authenticated gateway; never deploy APP_ENV=development publicly or send a shared manager token on behalf of all customers.

## How to add or change a module

1. State who may perform the action and which organization/resource the actor owns.
2. Define business states, rejection paths, idempotency and concurrency rules before UI work.
3. Add Postgres constraints and migration up/down pair when schema changes. Wire the new migration into Makefile/bootstrap and CI, with safe upgrade tests.
4. Add model -> service -> repository implementation in the relevant apps/api/internal domain. Keep money and authoritative state in the backend.
5. Add route permission/handler under internal/httpapi and matching contracts/*.openapi.yaml definitions. Test actor roles and cross-org ID tampering.
6. Add Next.js page/client/server action that forwards **current user's** token, not global service identity.
7. Add unit, integration, negative-access, replay and browser tests appropriate to the feature. Update docs and release checklist.
8. Verify staging with real OIDC, storage and provider stubs before promoting.

**Important distinctions:** a tenant record is not a login; a tenancy is not a lease; a payment is not an allocation; quote approval is not completion evidence; an agent proposal is not permission to execute.

## Developer troubleshooting

| Symptom | What to check |
| --- | --- |
| 401/403 | JWT issuer/audience/expiry, mapping, organization membership, role and linked owner/tenant scope |
| Web works but authenticated API fails | Forwarded per-request credentials, API address, production gateway trust model |
| Files return 503 | Storage bucket/credentials, S3 HEAD validation, document status |
| Push devices table missing | 000014 migration status and Makefile target, not just CI |
| Notification backlog grows | Running worker, provider configuration, attempts/retry/dead state |
| Provider paid event rejected | Correct raw-body signature, configured secret, event ID, tenant/organization and currency |
| Data present in wrong org | Stop rollout, inspect scoped SQL and authorization; never hide by broadening roles |

## Documentation map

- [Architecture and data flows](architecture.md)
- [Feature status audit](feature-status.md)
- [Implementation backlog](implementation-backlog.md)
- [API and integrations](api-and-integrations.md)
- [QA/launch gates](qa-and-release-checklist.md)
- [Ops and incident response](operations-runbook.md)
- [Domain model](domain-model.md), [accounting](accounting.md), [portals](portals.md)
- [Documents](document-storage.md), [notifications](notifications.md), [MCP](mcp.md)
- [Agent approvals](agentic-operations.md), [inspections](inspections.md)
- [Organization administration](organization-administration.md), [payments](payment-provider-integration.md)
- [Deployment](deployment.md), [local development](local-development.md), [product roadmap](product-strategy-and-roadmap.md)
