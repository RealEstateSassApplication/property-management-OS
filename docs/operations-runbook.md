# Operations, rollout and recovery

**Scope:** Document the existing reference deployment and required operator work, not a claim of production readiness. See also [deployment](deployment.md) and [production readiness](production-readiness.md).

## Topology

- Next.js web behind trusted HTTPS/OIDC gateway, Go API with readiness and pgx pool, independent Go notification worker, PostgreSQL, S3-compatible document bucket.
- MCP is a separately authorized stdio client process, not an openly accessible web endpoint.
- Local Compose runs Postgres 17; production-style Compose uses Postgres 18. **Do not reuse a database volume across major PostgreSQL versions** without a supported upgrade/dump-restore plan.
- Production config requires OIDC issuer/audience, storage bucket, safe secrets and non-log notification provider; PAYMENT_WEBHOOK_SECRET is required to enable paid-event ingress.

## Release sequence

1. Record candidate Git SHA, schema version and actor signing off. Review migration and backward compatibility.
2. Create verified DB backup **and independent document storage backup**, check retention/access and restoration.
3. Validate OIDC/issuer/audience, cookie and header trust model, TLS, S3, provider secrets, rate limits and secret management.
4. Apply new migrations once, in order, via an approved release job; never blindly rerun all migrations against existing DB. Migration version tracking remains an open P0.
5. Deploy API -> check /readyz, then worker, then web; MCP only when used by a trusted client.
6. Staging-smoke two-org identity/role access, rent/reversals/deposits, owner/tenant portal, documents, inspections, notifications and paid-event replay.
7. Monitor error and outbox/provider failure metrics, record release decision and rollback owner.

~~~bash
POSTGRES_PASSWORD=example docker compose -f docker-compose.production.yml config
~~~

This example checks Compose syntax, not live deployment. For image/backup/restore commands use [deployment](deployment.md).

## Alerts to add and validate

| Signal | Operator response |
| --- | --- |
| API /readyz, 5xx or p95 latency | Check DB pool saturation, deploy changes, logs/request IDs and ingress |
| Notification outbox pending/retry/dead age | Inspect worker/provider, do controlled idempotent replay after consent check |
| Payment provider signature failures/rejections/duplicates | Preserve fingerprints, pause adapter if necessary, reconcile with provider |
| S3 upload HEAD/download failure | Check bucket access, object lifecycle and data retention |
| Agent action failure/stale pending proposal | Require human review; never force direct ledger/quote update |
| Auth 401/403 abnormal trends | Review token/session gateway and revoked org/portal links |

These signals are targets; instrumentation, central alerting, on-call policies and thresholds must be configured in deployment.

## Backup and restore

scripts/backup-postgres.sh uses pg_dump custom archives and checks archive readability; scripts/restore-postgres.sh needs RESTORE_CONFIRM=YES because restoration is destructive. Restore to isolated staging first. Database dumps do **not** contain S3 file bytes. After restoring, validate identity isolation, schema/migrations, allocations, reversals, deposits, outbox behavior and authorized document availability. Record measured RPO and RTO.

## Incident handling

**Cross-org/PII leak:** disable exposed path, revoke credentials/grants, preserve audit/logs and identify affected scopes; escalate according to applicable privacy policy.

**Financial discrepancy:** stop suspect ingestion, preserve event IDs/fingerprints, reconcile against processor/bank. Use audited domain reversals/adjustments; never delete posted money records manually.

**Notification storm:** pause worker and isolate source of duplicate jobs. Check idempotency and recipient consent before controlled retry.

**Failed migration:** stop further incompatible deployments, compare backups/schema and choose reviewed roll-forward or restore. Do not attempt all down files blindly.

**Missing documents:** audit S3 settings/lifecycle, revoke compromised URLs and recover from independent object backup.

Assign named release operator, support owner, security escalation, finance reconciler, backup owner and customer incident contact before real data is accepted.
