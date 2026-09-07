# Local development

A fresh clone of Property Management OS can run locally without configuring production OIDC, S3-compatible storage, a notification provider, or a payment gateway.

## Prerequisites

Install:

- Docker with Docker Compose
- Go 1.27.1 or compatible Go 1.27 toolchain
- Node.js 24 with npm
- `make`

## First run

From the repository root:

```bash
make dev-setup
make dev
```

`make dev-setup` will:

1. create `.env` from `.env.example` when it does not exist;
2. start the local PostgreSQL container;
3. wait for PostgreSQL readiness;
4. apply every migration and load deterministic development data on a fresh database;
5. download the Go module graph; and
6. install the Next.js workspace dependencies.

`make dev` starts the API, notification worker, and Next.js web application together and shuts them down together on Ctrl-C.

Local endpoints:

```text
Web           http://localhost:3000
API           http://localhost:8080
Liveness      http://localhost:8080/healthz
Readiness     http://localhost:8080/readyz
```

The seeded manager identity is:

```text
Organization  11111111-1111-1111-1111-111111111111
User          22222222-2222-2222-2222-222222222222
Role          admin
```

The seeded owner and tenant portal identities are already included in `.env.example`.

## Optional local services

Document uploads/downloads return a service-unavailable response when no S3-compatible storage is configured. All other document metadata behavior remains available.

The notification worker defaults to `NOTIFICATION_PROVIDER=log`, so deliveries appear in the local process output rather than contacting an external service.

The payment-provider webhook remains disabled when `PAYMENT_WEBHOOK_SECRET` is blank.

Production OIDC is not required when `APP_ENV=development`; the server-side development user IDs in `.env` are used by the local Next.js application and MCP client.

## MCP

The normal `make dev` command does not start MCP because MCP uses stdio and is normally launched by an MCP client. To run it directly:

```bash
set -a && . ./.env && set +a
make mcp
```

## Stop local infrastructure

Stop the application processes with Ctrl-C. Stop PostgreSQL with:

```bash
make dev-stop
```

The PostgreSQL named volume is retained. `make dev-setup` detects an existing initialized schema and will not replay first-run migrations or seed data.

For an existing development database after pulling new migrations, apply the new migration target(s) explicitly before starting the application.
