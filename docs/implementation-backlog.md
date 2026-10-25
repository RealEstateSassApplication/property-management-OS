# Remaining work: prioritized engineering backlog

**Evidence snapshot:** 2026-10-25, main 6df541f. P0 = required before real users/money; P1 = first paid-pilot product differentiation; P2 = validated later expansion. Tasks are incomplete features, verified defects or unverified launch gates, **not** all confirmed code bugs.

## P0: correctness, access control and safe launch

### P0-01 — Migration parity: confirmed local setup bug
- The repository has migrations/000014_push_devices.up.sql and down file, but Makefile previously ended at migration 000013. CI applies all numbered SQL files using a glob, masking local bootstrap drift.
- **Fix in this PR:** add migrate-push-devices / migrate-push-devices-down targets and include in migrate-all and reverse-order migrate-down; add CI parity check.
- **Acceptance:** fresh setup has push_devices; previous-version database applies *only* unapplied 000014; no accidental re-running non-idempotent SQL.

### P0-02 — Introduce tracked and safe schema upgrades
- Use a durable migration version ledger/tool, ordered up/down policy, repeat deployment/partially applied migration checks and production restore plan.
- **Acceptance:** startup and releases cannot silently skip schema revisions just because organizations table exists; upgrade from previous release passes with existing real-like data.

### P0-03 — Finish browser identity and sessions
- Build secure OIDC authorization-code/PKCE or independently audited trusted gateway flow, session cookie/token custody, refresh/logout, invitations and organization switching.
- **Acceptance:** wrong/missing/expired token, forged identity headers, CSRF, role change, revoked account and cross-organization access fail closed. No shared admin token in production Next.js.

### P0-04 — Tenant isolation, document and agent security audit
- Automated tests for each role/org, foreign resource IDs, effective-dated owner access, tenant portal link revocation, signed document URLs, uploads and arbitrary document access.
- Attack-test MCP prompt injection, authorization bypass, stale agent quote context, webhook replay/signature and PII/log exposure.
- **Acceptance:** human-reviewed negative-test matrix and no unresolved cross-organization/data-leak blocker.

### P0-05 — Finance and provider reconciliation
- Exercise concurrent allocations, duplicate events, void payments, adjustments, deposits, ownership rounding, wrong currency, provider retries and settlement exceptions.
- Add true provider-specific verification/checkout adapters *before* production, including dispute/refund/chargeback and bank settlement matching processes. Generic signed event endpoint alone is not a complete gateway.
- **Acceptance:** accountant-signoff, no unexplained ledger/provider differences, explicit auditable corrections rather than SQL deletion.

### P0-06 — Deployment/operations readiness
- Real OIDC, S3, notification delivery and payment stubs in staging; TLS/rate limiting, secret rotation, backup verification and object retention.
- Add meaningful tracing, 5xx/latency and worker/dead-letter/provider alerting, tested restore, incident ownership, RPO/RTO and rollback rehearsals.
- **Acceptance:** recovery drill, two-org E2E staging suite, go/no-go document; no public production launch based on green CI alone.

## P1: differentiated paid-pilot product

### P1-01 — Switch Kit Phase 2 ([issue #20](https://github.com/RealEstateSassApplication/property-management-OS/issues/20))
- Keep existing CSV preview read-only. Build org-scoped staged import jobs, source hash, existing-record collision checks, manager-approved create/link/skip actions, idempotent apply, resumable row outcomes and reconciliation/rollback report.
- Import properties and units first, then controlled owners/tenants/leases and separately audited historic money.
- **Acceptance:** same file imported twice creates no duplicates; forgery, oversize, wrong-org and conflicts rejected; counts reconcile.

### P1-02 — Owner Proofbook ([issue #19](https://github.com/RealEstateSassApplication/property-management-OS/issues/19))
- Generate read-only scoped evidence backed by posted receipts/allocations, adjustments, expenses, effective ownership, quote approvals, completion evidence, inspections and audit events. Add versioned PDF/export, redacted short-lived shares.
- **Acceptance:** every amount references records with proper currency/reversal handling; revoked owner never sees another property; distinguish missing evidence from completed work.

### P1-03 — Optional Avara distribution contract ([issue #21](https://github.com/RealEstateSassApplication/property-management-OS/issues/21))
- Explicit consent, organization/account and external property IDs, signed versioned event contract, idempotent outbox/consumer, dedupe, replay and manual conflict resolution.
- Begin with one-way applicant handoff to manager review; do not create leases or financial postings automatically.
- **Acceptance:** duplicate/forged/reordered events or revoked consent cannot overwrite/duplicate records; Property OS keeps ledger authority.

### P1-04 — End-to-end manager/tenant/owner usability
- Complete first-run organization onboarding, missing/error/loading states, evidence and mobile-friendly field UX, notification preferences/provider-backed deliverability, operational exports and support workflow.
- **Acceptance:** three real pilot managers can enter/verify a portfolio, run a month's rent and maintenance, deliver accurate owner reporting without developer SQL fixes.

### P1-05 — SaaS commercial lifecycle
- Secure company signup/invite, plan usage/billing, entitlement, subscription cancellation, customer export/deletion, support access and legal/privacy/retention policies.
- **Acceptance:** complete tenant provisioning/deprovisioning, invoice/trial/cancellation paths, permissions and data export validated.

## P2: prioritize from evidence and pilot demand

- Regional LKR/bank transfer matching and reporting, only where customers need it.
- Scheduled arrears and renewals with consent, dedupe, human approval and policy controls.
- Independently verified field/mobile offline workflows and push delivery.
- More complete telemetry/SLOs, advanced analytics, anomaly detection and performance profiling.
- Hospitality inventory/OTA booking engine only as a **separate later** domain; do not equate an Airbnb reservation with a long-term lease.

## Recommended PR slices and gates

1. Migration parity (this PR) -> versioned runner/upgrade tests.
2. Secure login/session and org switching -> role/tenant isolation matrix.
3. Payment/provider reconciliation -> staging recovery/monitoring.
4. Approved property/unit importer -> owner/tenant/lease extensions.
5. Proofbook scoped read API -> UI/report exports.
6. Avara consent and mapping -> signed events/replay/monitoring.
7. Pilot UX and billing lifecycle -> measurable customer adoption.

Every PR must document domain/permission contracts, concurrency/idempotency (where needed), exact tests, backwards compatibility and staging evidence. Update [feature inventory](feature-status.md) after merge.
