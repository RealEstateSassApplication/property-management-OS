# Production readiness

This document separates feature completeness from operating the platform safely with real users and real money.

## Required runtime configuration

Production (`APP_ENV` other than `development`/`test`) already requires OIDC and S3-compatible document storage. Before enabling payment ingestion, configure a high-entropy `PAYMENT_WEBHOOK_SECRET` and rotate it through the deployment secret manager rather than committing it to source control.

Minimum production configuration:

- PostgreSQL using encrypted connections and provider-managed backups;
- `OIDC_ISSUER_URL` and `OIDC_AUDIENCE`;
- S3-compatible `STORAGE_BUCKET` and region/credentials;
- production notification provider configuration;
- `PAYMENT_WEBHOOK_SECRET` when payment provider ingestion is enabled;
- HTTPS termination at the ingress/load balancer;
- API and worker replicas sized independently.

## Health and readiness

`GET /healthz` is a process/liveness signal. `GET /readyz` performs a bounded PostgreSQL ping and returns `503` when the API should not receive business traffic. Deployment platforms should use `/readyz` for readiness and `/healthz` for liveness.

## Security baseline

The API sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `Cache-Control: no-store`, and a restrictive CSP. Business routes remain protected by OIDC identity plus server-side organization membership/RBAC. Public payment callbacks use provider signature authentication instead of user identity.

Additional edge controls should be configured at the ingress/API gateway:

- TLS 1.2+;
- request and connection rate limits;
- IP reputation/WAF rules where appropriate;
- request body limits consistent with application limits;
- access logs with secrets/tokens redacted;
- webhook-specific rate limits that tolerate legitimate provider retries.

## Database and migrations

Run schema migrations once per deployment through a controlled release job before starting the new application version. Never let every API replica race migrations at startup. Take a verified backup before destructive/schema-sensitive releases.

Required operational exercises before launch:

1. restore the latest production-like backup into an isolated database;
2. run all migrations on the restored copy;
3. verify organization, rent, accounting, inspection and provider-event invariants;
4. document RPO/RTO targets and escalation ownership.

## Payment operations

Provider events are idempotent by `(provider,event_id)` and retain their payload fingerprint. Operators should monitor rejected provider events and reconcile provider settlement totals against Property OS posted payments. Refunds/chargebacks should use explicit reversal workflows rather than deleting payment history.

## Observability still expected

The codebase should emit OpenTelemetry traces/metrics and structured logs to the production telemetry stack. At minimum alert on:

- API readiness failures;
- PostgreSQL saturation/connection failures;
- notification worker backlog and terminal failures;
- payment-provider rejected events and signature failures;
- agent-action execution failures;
- elevated 5xx rate/latency;
- object-storage errors.

## Deployment gate

A release is production-ready only when:

- Go tests/build pass;
- PostgreSQL migrations and invariant checks pass;
- Next.js lint/build pass;
- production secrets are configured outside source control;
- backup restore has been tested;
- OIDC, storage, notifications and payments have been exercised in a staging environment;
- rollback and incident procedures are documented.
