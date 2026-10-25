# API and integration reference map

**Authoritative source:** apps/api/internal/httpapi/router.go, inspection_routes.go and production_routes.go; data contracts live in contracts/*.openapi.yaml. The routes below are representative. Refer to Go handlers and specs for exact payloads/errors.

Most requests use /api/v1 prefix, verified Authorization: Bearer token plus X-Organization-ID. Local/test only: X-User-ID is a dev fallback. The organization header selects scope but **does not grant membership**.

| Area | HTTP route examples | Contract |
| --- | --- | --- |
| Portfolio | GET/POST /properties, GET/PATCH /properties/{propertyID}, units under properties and /units/{unitID} | contracts/openapi.yaml |
| People, occupancy, leasing | GET/POST /tenants, /tenancies, /leases plus detail GET/PATCH | contracts/openapi.yaml |
| Owners | GET/POST /owners and /ownership-interests | contracts/openapi.yaml |
| Rent | GET/POST /rent/obligations and /rent/payments; POST /rent/allocations | contracts/openapi.yaml |
| Accounting | /accounting/rent-adjustments, payment reverse, security deposits, expenses, owner statements | contracts/accounting.openapi.yaml |
| Maintenance | /maintenance/vendors, /requests, /work-orders, /quotes, /evidence and transition/approval routes | contracts/openapi.yaml |
| Inspections | /inspections, items, completion, acknowledge | contracts/inspections.openapi.yaml |
| Documents | /documents, uploads, complete, download, DELETE | contracts/documents.openapi.yaml |
| Notifications/push | /notifications, /notifications/rent-reminders, /notifications/push, /push-devices | contracts/notifications.openapi.yaml |
| AI approvals | /agent-actions, /agent-actions/maintenance-quote-approvals, /agent-actions/{actionID}/decision | contracts/agent-actions.openapi.yaml |
| Owner/tenant portal | GET /portal/owner and /portal/tenant; tenant maintenance POST | contracts/portals.openapi.yaml |
| Org admin/reporting | /organization, /organization/members, /organization/portal-links, /reporting/dashboard, /audit-events | contracts/admin-reporting.openapi.yaml |
| Onboarding | POST /onboarding/properties/preview (text/csv, 512 KiB/500 rows) | contracts/onboarding.openapi.yaml |
| Payment integration | POST /integrations/payments/generic-hmac/webhook | contracts/payment-webhook.openapi.yaml |

## Verified domain flows

**New rent:** create property/unit, register tenant, tenancy, active lease, monthly obligation derived from lease, posted payment and authorized allocation. Payment does **not** automatically reduce outstanding obligation until allocated; do not treat posted cash and allocation as equivalent.

**Maintenance:** request -> work order -> vendor quote -> cost-approval role -> execution -> completion evidence -> completed work order/request. Approval is not completion.

**Documents:** authorized metadata/upload initiation -> browser presigned upload -> server HEAD verification -> available/quarantined -> separately authorized presigned download.

**Notifications:** API commits durable outbox row; worker claims and retries deliveries. Same-day rent reminder keys limit duplicates. Delivery depends on provider configuration.

**MCP:** server (stdio) calls HTTP APIs, never databases. Read tools include portfolio, arrears, leases, maintenance and quote/proposal listings. Controlled writes include reminders, maintenance requests and human-reviewed quote proposals.

## External ownership and security

| Boundary | What exists | Not implemented/proven |
| --- | --- | --- |
| Generic payment HMAC | Signed raw-body ingestion of normalized payment.paid, idempotent provider-event ID and audit | Native provider payment checkout, merchant settlement match, refunds/chargebacks |
| Notifications | Development log, webhook adapter and Expo delivery path | End-to-end production delivery templates, subscriptions, consent and receipts |
| Avara | Strategic/API boundary and external ID concept | Opt-in mapping, signed outbox, conflict/replay handling and real one-way approved application transfer |
| OpenAPI | Split contracts checked into repository | Full generated aggregate API/SDK or demonstrated contract-to-handler coverage |

Do not use client-computed balances, expose cross-org foreign identifiers, log secrets/presigned URLs or let partner applications write financial tables. Add contract and negative-path tests for every new mutation.

## Local request

~~~bash
curl -fsS http://localhost:8080/api/v1/reporting/dashboard \
  -H 'X-Organization-ID: 11111111-1111-1111-1111-111111111111' \
  -H 'X-User-ID: 22222222-2222-2222-2222-222222222222'
~~~

Use real OIDC access tokens in production, not the sample identity.
