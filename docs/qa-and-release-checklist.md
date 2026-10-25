# QA plan and production go/no-go checklist

A repository containing code, tests and CI is not equivalent to a customer-tested SaaS. Add these cases as automated suites and record staging outcomes for a real release SHA.

## Baseline automated commands

~~~bash
make test-api
make build-api
npm run lint:web
npm run test:web
npm run build:web
~~~

CI currently applies SQL migrations by filename glob and checks deterministic fixtures. Verify the **Makefile local bootstrap path** as a separate parity test. Native browser E2E and external provider testing need expansion.

## Risk-based test matrix

| ID | Scenario and expected result |
| --- | --- |
| AUTH-01 | Valid OIDC token resolves issuer+subject; invalid signature/issuer/audience/expiry rejected |
| AUTH-02 | Org A user cannot access Org B record by swapping path ID or X-Organization-ID |
| AUTH-03 | Owner/tenant after unlink or role change cannot access unrelated property, finance or document |
| AUTH-04 | Real browser login, org picker, expired session, refresh/logout and CSRF covered |
| TEN-01 | Two concurrent active tenancies cannot occupy one unit; lease lifecycle transitions validated |
| RENT-01 | Concurrent allocations cannot overallocate payment or obligation; partial payment works |
| RENT-02 | Currency mismatch, void receipt and reversal correctly reopen receivables |
| ACC-01 | Deposit transaction cannot overspend liability; owner share and rounding reconcile by effective date |
| PAY-01 | Missing/tampered HMAC rejected; identical event replay is idempotent; changed payload with reused ID rejected |
| PAY-02 | Real provider settlement/refund/dispute totals reconcile or produce explicit exceptions |
| MAIN-01 | Unauthorized quote approval rejected; one quote selected; work completion needs evidence |
| INSP-01 | Cross-org evidence doc rejected; completed inspection immutable |
| DOC-01 | Wrong-org file access, bad MIME/size, head mismatch/quarantine, short-lived URLs verified |
| NOTIF-01 | Concurrent workers cannot claim same outbox item; failures retry/dead, duplicate reminder prevented |
| PUSH-01 | Fresh and upgraded DB contain push_devices and own-device RBAC works |
| AGENT-01 | MCP agent cannot post payments or approve quotes; stale proposal reevaluated by domain |
| CSV-01 | Malformed/oversize CSV rejected; preview performs zero database writes |
| CSV-02 | Future approved import dedupes existing records and supports safe replay/reconciliation |
| AVARA-01 | Future forged, duplicate, reordered, revoked events create no silent overrides |
| OPS-01 | Isolated DB+object restore succeeds with schema integrity and authorized file availability |
| E2E-01 | Manager sets up property/unit, tenant, lease, rent, payment, maintenance and statement |
| E2E-02 | Tenant creates scoped maintenance request; owner sees only authorized property evidence |

Build test fixtures for **two organizations**, two owners with changing interests, multiple roles, duplicates, and concurrency. Financial and authorization negative paths require actual database/integration checks, not only mocks.

## Final sign-off

- [ ] API and frontend checks green for exact SHA
- [ ] Fresh local setup and upgrade of previous database schema tested
- [ ] Production-like OIDC and request identity sessions tested, including owner/tenant roles
- [ ] Cross-org/ID tampering and document access security cases passed
- [ ] Money provider events/settlements and owner/deposit balances reconciled by a human reviewer
- [ ] Staging S3, notification provider, push and webhook failure paths passed
- [ ] DB **and object storage** backup/restore drill measured; RPO/RTO published
- [ ] TLS, rate limits, secret rotation, incident owners and privacy/retention confirmed
- [ ] Pilot operators signed off on workflows and known limits documented
- [ ] Rollback/deployment procedure and customer support escalation ready

Any failed P0 means remain in development or controlled pilot; do not claim production-grade security or full accounting correctness.
