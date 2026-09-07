# Deployment and recovery

Property Management OS ships separate production images for the web application and Go runtime targets. The API, worker and MCP use `apps/api/Dockerfile` with a `TARGET` build argument; the web application uses `apps/web/Dockerfile` and Next.js standalone output.

## Build images

```sh
docker build -f apps/api/Dockerfile --build-arg TARGET=api -t property-os-api:local apps/api
docker build -f apps/api/Dockerfile --build-arg TARGET=worker -t property-os-worker:local apps/api
docker build -f apps/api/Dockerfile --build-arg TARGET=mcp -t property-os-mcp:local apps/api
docker build -f apps/web/Dockerfile -t property-os-web:local .
```

The Go runtime images run as a non-root user. The web image also runs as a non-root user and contains only the Next.js standalone server/runtime artifacts needed to serve the application.

## Production-style Compose

`docker-compose.production.yml` is intended as a self-hosted/reference topology rather than a substitute for a managed production database. Set real secrets in an uncommitted `.env`, especially `POSTGRES_PASSWORD`, OIDC values, storage credentials, notification credentials and `PAYMENT_WEBHOOK_SECRET`.

Validate configuration before deployment:

```sh
POSTGRES_PASSWORD=example docker compose -f docker-compose.production.yml config
```

Start the long-running services:

```sh
docker compose -f docker-compose.production.yml up -d postgres api worker web
```

Run the MCP server only when a stdio client needs it:

```sh
docker compose -f docker-compose.production.yml --profile mcp run --rm mcp
```

The API exposes `/healthz` for liveness and `/readyz` for readiness. The latter verifies PostgreSQL connectivity and is used by dependent services.

## Release sequence

For a controlled production deployment:

1. create and verify a database backup;
2. deploy/run migrations once through a release job;
3. start the new API and wait for `/readyz`;
4. start/roll the worker;
5. deploy the web application;
6. run smoke tests for authentication, portfolio, rent/accounting, documents, maintenance, inspections and payment webhook ingestion;
7. observe error rate, payment-provider rejections and notification backlog before completing the release.

Do not run migrations independently from every API replica.

## Backup

The repository includes a PostgreSQL custom-format backup helper:

```sh
DATABASE_URL='postgres://...' sh scripts/backup-postgres.sh
```

Use `BACKUP_DIR` or `BACKUP_FILE` to control the destination. The script runs `pg_restore --list` after `pg_dump` so a completely unreadable archive is detected immediately. This is an archive-structure check, not a substitute for a restore drill.

## Restore drill

Restore only into an isolated/staging target unless an incident procedure explicitly authorizes a production restore. The restore helper requires an explicit destructive confirmation:

```sh
RESTORE_CONFIRM=YES \
DATABASE_URL='postgres://isolated-target/...' \
sh scripts/restore-postgres.sh backups/property-os-YYYYMMDDTHHMMSSZ.dump
```

After restore, apply any migrations newer than the backup and run the CI/database invariants against the restored environment before serving traffic.

## Request correlation

Every API response includes `X-Request-ID`. A valid incoming request ID is preserved so ingress/proxy logs can correlate with application logs; otherwise the API generates one. Structured access logs include request ID, method, path, status, response bytes, duration and remote IP. Query strings and request bodies are intentionally excluded from access logs.

Production ingress should preserve/forward `X-Request-ID`, redact authorization headers and secrets, enforce TLS/rate limits and export logs to the central telemetry platform.
