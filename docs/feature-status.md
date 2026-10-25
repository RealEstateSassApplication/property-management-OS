# Feature status: implemented, partial and planned

**As of 2026-10-25, main 6df541f.** Source-code inventory, not a staging certification. “Present” means relevant code/routes/schema are in the repository, not that a real customer environment has passed acceptance tests.

| Capability | Code status | Evidence / remaining validation |
| --- | --- | --- |
| Multi-org API, OIDC verification, RBAC | Present foundation | internal/auth and httpapi/context.go; native browser sign-in/logout/org switcher unfinished |
| Properties and units | Present | internal/properties, units, web/properties |
| Tenants, tenancies, leases | Present | internal/tenants, tenancies, leases, web/tenants and leases |
| Owners and effective-dated interests | Present | internal/owners, migrations/000003 |
| Rent obligations, posted payments and allocations | Present | internal/rent, web/rent |
| Adjustments, reversals, deposits, expenses, owner statements | Present foundation | internal/accounting; external bank/provider reconciliation unfinished |
| Maintenance, vendors, quotes, work orders, completion evidence | Present | internal/maintenance and web/maintenance |
| Inspections and acknowledgements | Present | internal/inspections, inspection_routes.go |
| Documents and presigned object storage | Present foundation | internal/documents; staging S3/retention review necessary |
| Durable outbox, notification worker, push endpoints | Present; local schema fix needed | internal/notifications; migration 000014 absent from previous Makefile migrate-all |
| Owner/tenant resource-scoped summary portals | Present foundation | internal/portals, web/owner-portal, web/tenant-portal |
| Organization settings, member administration and audit/reporting | Present | internal/orgadmin, reporting, web/settings, web/reports |
| AI MCP, constrained actions, human quote-approval review | Present foundation | cmd/mcp, internal/agentactions, web/agent-actions |
| Generic HMAC payment-paid webhook | Present adapter | internal/paymentproviders; not a complete native gateway |
| CSV property onboarding preview | Present read-only | internal/onboarding/property_csv.go; no data is imported |
| Native web login/session/refresh/logout | Partial/missing UX | production currently requires token-forwarding trusted ingress |
| Switch Kit import staging/approval/dedupe/rollback | Planned | Issue #20 |
| Owner Proofbook provenance/export | Planned | Issue #19 |
| Avara mapping, signed handoff/event replay | Planned | Issue #21 |
| Actual provider settlement, refunds/chargebacks, bank matching | Planned | domain reversal foundation is not full external reconciliation |
| Mobile companion app | Outside this repository | mentioned in roadmap; independently assess its code/devices/releases |
| Full production security, reliability, finance and UAT sign-off | Unverified | CI and source files cannot establish real-world readiness |

## Source and PR facts

Main includes [PR #18](https://github.com/RealEstateSassApplication/property-management-OS/pull/18) (CSV preflight + roadmap; merged Oct 10) and [PR #17](https://github.com/RealEstateSassApplication/property-management-OS/pull/17) (push backend; merged Sep 12). [PR #22](https://github.com/RealEstateSassApplication/property-management-OS/pull/22) is closed **without merging** and concerns an isolated RevenueLeak OS pilot, not this product baseline. No open PRs were returned by the GitHub query at the time of review. Open Issues #19, #20 and #21 define substantive planned work.

## Evidence limitations and update rule

This review inspected the file tree, relevant handlers/services, Makefile, local setup, env examples, source docs, migration files, workflow and PR metadata. It did **not** execute Go, Next.js, PostgreSQL, provider sandboxes, a live OIDC login or a real backup/restore. Keep “Present”, “Planned”, “Unverified” separate. Update this table for each merged slice; never market “production-ready”, “verified evidence” or “AI-approved” before testing and authorization review.
