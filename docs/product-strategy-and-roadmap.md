# Avara + Property OS: product strategy and delivery roadmap

**Updated:** 2026-10-08  
**Status:** Strategic plan and release criteria; not a claim that every feature is production-ready.

## North star

**Avara Property OS helps independent property managers show exactly what happened to every property, payment and maintenance job, while Avara Marketplace optionally helps those managers find tenants and guests.**

Do not build a generic classifieds portal or a visual imitation of Airbnb. The product wins by reducing paperwork, proving work was completed, explaining money, and making a property manager's existing data portable.

### Two products, one promise

1. **Property OS (B2B system of record):** properties and units, organizations, owners, occupancies, leasing, receivables, immutable payment allocations, deposits, approvals, inspections, vendor operations, documents and audit history.
2. **Avara Marketplace (customer acquisition and transactions):** public listings, inquiries, applications, and eventually short-stay bookings. It must remain optional for Property OS customers.
3. **Mobile (field-work client):** staff capture inspection findings and maintenance completion evidence at the property. It never becomes an independent accounting source of truth.

**Architectural rule:** Property OS owns its domain records in PostgreSQL. Avara's MongoDB owns marketplace concerns. Exchange authorized API calls/events using explicit external IDs and idempotency. Never dual-write property accounting, read across databases, or synchronize financial totals as if they were authoritative.

## The competitive choice

Initial customer profile: **property management businesses managing roughly 20–150 residential units**, especially those coordinating multiple property owners, collecting rent, and managing maintenance using Excel, WhatsApp and email.

Research targets (not endorsements): [Buildium](https://www.buildium.com/), [AppFolio](https://www.appfolio.com/), [DoorLoop](https://www.doorloop.com/) in residential management; [Guesty](https://www.guesty.com/), [Hostaway](https://www.hostaway.com/) and [Hospitable](https://hospitable.com/) in hospitality.

What we will **not** claim: that an agent, dashboard, portal, rental listing, basic channel sync, or a lower price alone is unique. These are established categories. We need measurable improvements, not feature-count marketing.

### Three differentiators to validate in real pilots

**1. Owner Proofbook (planned): transparent operations, not dashboard decoration.**

For each property and reporting period, create a shareable *permission-scoped* evidence trail: rent assessed and allocated, arrears, vendor quote and human approval, work order, inspection photos, completed evidence, reversals, owner statement, and exact provenance. It should answer "What happened? Who approved it? Where did the money go? What proof exists?" without exposing unrelated tenants or properties. Export as PDF with event/evidence identifiers; cryptographic hashes and tamper-evident history are later hardening tasks, **not** already delivered guarantees.

**2. Switch Kit (preflight implemented in this tranche): minimize switching pain.**

Import from structured CSV/Excel exports through validate → preview → duplicate detection → authorized approval → idempotent import → reconciliation. The **first implemented slice** is *read-only property CSV preflight*. It does not create data. A full import job with rollback/reconciliation comes later. This should become an important distribution tool, including property ID mapping from Avara.

**3. Human-controlled Operations Copilot (foundation exists):**

Use the existing constrained MCP/API access to summarize arrears, lease expiries, maintenance and outstanding owner issues, draft reminders, and propose approvals. High-risk financial and legal operations require explicit human approval, up-to-date domain checks, and audit events. Never let an LLM approve its own payout, alter the ledger, or send unrestricted private documents.

### Why this wedge is worth testing

Existing vendors offer accounting, portals, automation and maintenance workflows. Our hypothesis is that smaller multi-owner managers will pay for **verifiable owner communication + smoother migration + regional workflows**. This hypothesis is unproven until prospective customers use it and pay. Avoid unfounded claims about competitor gaps or guaranteed savings.

## Currently implemented vs. planned

| Capability | Current state from code/repository | Productization requirement |
| --- | --- | --- |
| Organization-scoped Go API, PostgreSQL, OIDC/RBAC | Implemented foundation | Security review, real-identity staging tests, abuse testing |
| Properties, units, owners, tenants, leasing | Implemented | More migration and lifecycle coverage |
| Rent ledger, partial allocations, deposits, adjustments and owner statements | Implemented foundation | Bank settlement reconciliation, exception handling, accountant sign-off |
| Work orders, approvals, inspections, documents, notification outbox | Implemented foundation | Vendor/customer usability, reliable provider delivery and operational monitoring |
| Agent proposals and human approvals | Implemented for constrained workflows | Evaluate prompt-injection, permission bypass and stale-state behavior |
| Native companion | Implemented in separate repository | Real device E2E, production push, release/signing/privacy |
| **Property CSV preflight** | **Included in this first build slice** | Review results with pilot operators; build real imports later |
| Owner Proofbook with export and verified evidence links | **Planned** | Build from authorized immutable records; verify inclusion/exclusion rules |
| Avara ↔ Property OS integration | **Planned**, contract boundary documented | Consent, tenant/org mapping, idempotent outbox, replay and reconciliation |
| Hospitality PMS / OTA channel sync / guest bookings | **Future**, separate business workflows | Inventory holds, timezones, payments/refunds, channel SLAs, housekeeping |

## Phased roadmap and nonnegotiable exit criteria

### Phase 0 — Safety and release discipline (blocking)

- Verify mainline CI (Go tests/build/gofmt, Next lint/build, PostgreSQL migrations and invariants).
- Exercise real OIDC sessions, organization isolation, owner/tenant access, secret rotation, backups/restores, and incident response in staging.
- Exercise payment webhook signature, idempotency, wrong-currency and duplicate-delivery scenarios.
- Establish application error reporting, database metrics, notification failure alerts, request correlation and a documented roll-back path.
- **Exit:** no unresolved P0 security, data-loss, access-control or finance-reconciliation defects; written staging sign-off. CI green is necessary but not sufficient.

### Phase 1 — Switch Kit and fast onboarding (current slice starts here)

- **Now:** CSV-only property preflight with 512 KiB/500-row cap, normalization, duplicate checks, row-level errors, strict org permission, zero database writes.
- **Next:** template mapping for major export formats, reversible import job tables, idempotency keys, import approval by admin/manager, existing-record collision detection and reconciliation reports.
- **Exit:** three pilot operator CSVs can be validated safely; later, real imports require no silent overwrites and can be reconciled to source counts.

### Phase 2 — Owner Proofbook (primary customer differentiator)

- Event links across rent, allocations, owner statements, maintenance quotes, vendor approvals, photos, inspections and audit logs.
- Owner-facing monthly report and downloadable evidence-backed statement, redaction and expiring share grants.
- Reports only include records within verified organization + owner property scope.
- **Exit:** every displayed amount reconciles to posted records; no unrelated tenant information leaks; owners can trace each issue and payment to source events.

### Phase 3 — Avara distribution adapter

- Opt-in listing/application sync with explicit property/organization identity mapping; Avara remains optional.
- Outbox-driven, signed event delivery, deduplication, retries, failure visibility and replay. Start **one-way** Avara approved tenancy/application → Property OS manager review.
- Reserve marketplace booking integration for the hospitality module; do not turn a lease into a short-stay booking.
- **Exit:** retries do not create duplicate properties/tenancies; manual correction and audit trail exist.

### Phase 4 — Regional operating advantage

- Bank transfer matching, LKR reports, messaging preferences, statement distribution and Excel workflows where customers actually request them.
- First-class staff and vendor scheduling, mobile inspections and recorded completion evidence.
- **Exit:** measurable reduction in monthly manager administrative time and owner-support questions across pilot customers.

### Phase 5 — Hospitality (only after residential validation)

- Separate booking reservation and inventory engine with timezone-safe availability, expiring holds, refunds, cancellations, fees and reconciliation.
- iCal can be an initial connection but is **not real-time inventory assurance**. Evaluate official channel integrations subject to partner access.
- Cleaning turnovers, guest messaging, operator payouts and owner profitability; marketplace can offer incremental bookings.
- **Exit:** demonstrated reliability under parallel reservations, failed payments, provider redelivery, and cancellations. No launch if double-booking risk remains.

## Pilot validation and commercial gates

Interview at least **15–20** managers and ask for real examples of their workflows, invoices and reports (with consent and redaction). Offer three paid or contractually committed pilots; validate alternatives, switching costs, willingness to pay and required support. Pilot metrics:

- **Data integrity:** 0 unexplained reconciliation differences; 0 cross-organization exposures; 0 duplicate finance postings.
- **Operations:** time to onboard a 50-property portfolio; weekly staff hours saved; percentage of maintenance jobs with adequate evidence.
- **Owner trust:** number of owner queries about missing proof or unclear charges; time to deliver an accurate statement.
- **Commercial:** paid conversion, activation, retained properties, support burden and gross margin. Avoid promising fixed pricing or a $1M ARR outcome before validation.

Pricing hypothesis only: test low-friction base subscription plus per-managed-unit tiers and paid onboarding; do not claim this beats every competitor's cost.

## Security, product and accuracy guardrails

- Do not market "verified", "insured", "protected", "AI-approved", "instant sync", or "production-ready" without evidence and actual operations backing the claims.
- Per-tenant/per-owner permissions apply to files, statement totals, signed downloads and generated narratives.
- Payments and security deposits are liabilities/transactions, not dashboard guesses. Reversals must remain auditable.
- AI only reads authorized data and proposes actions according to least privilege; consequences are never silently executed.
- Prefer server-authoritative values, explicit financial currency and minor units, and idempotent workflows.
- Minimize personal information in import previews, logs and support artifacts.
- No unreviewed production deployment or merging failing checks.

## First build slice: property CSV preflight

- UI: `/onboarding` in Property OS web.
- API: `POST /api/v1/onboarding/properties/preview`, `Content-Type: text/csv`, organization membership with `portfolio:manage`.
- Expected columns: `referenceCode,name,propertyType,addressLine1,city,region,countryCode,externalAvaraPropertyId`. Required: `name,propertyType,addressLine1,city,countryCode`.
- Normalizes property type to lowercase and country code to uppercase; accepts optional BOM and quoted commas.
- Returns `totalRows`, `readyRows`, `invalidRows`, at most 30 valid `previewRows`, `issues`, and `canImport` **meaning file passes this preflight only**.
- Rejects duplicate reference codes and duplicate Avara IDs **within the file**. Checking existing database records is explicitly reserved for the actual import workflow.
- A successful preview **does not persist data**. The manager must still create records manually until the reviewed import tranche ships. Avoid uploading files with irrelevant personal information.

Sample: [property template](samples/properties-onboarding.csv).

### Why not combine the databases?

Property OS's rent/finance domain includes immutable records and organization permissions; Avara has marketplace-focused booking and listing models. Combining these prematurely could introduce dual-ledger discrepancies and security leaks. Connect via externally visible contracts rather than copying tables or encouraging the two systems to write the same fields.

## Definition of done for every milestone

1. Domain rules, API schema and authorization rules documented.
2. Unit tests for valid, invalid, duplicate, concurrent and unauthorized behaviors as applicable.
3. Integration tests with staging PostgreSQL/provider stubs for state-changing operations.
4. Accessibility and small-screen UX reviewed, with readable actionable errors.
5. CI build, lint, formatting, migrations and rollback checks green.
6. Negative-path and security review completed; operators understand what is not yet supported.
7. Real pilot feedback, product metrics and operational ownership recorded before launch.

A roadmap is a sequence of testable hypotheses and release gates, not a guarantee of bug-free software.
