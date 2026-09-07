# Organization administration and reporting

This tranche makes Property Management OS self-administerable by a property-management company instead of requiring direct database changes.

## Authorization model

- `admin`: full organization settings, membership, portal links, reporting and audit access.
- `manager`: operating settings, portal links, reporting and audit access, but cannot add/remove/change organization memberships.
- `accountant`: reporting and audit visibility in addition to existing financial permissions.
- `viewer`: reporting visibility without organization administration.
- `owner`, `tenant`, and `agent`: no organization administration access.

The organization service refuses to remove or demote the final administrator.

## Organization settings

`organization_settings` stores operating defaults such as timezone, default currency, country code and billing email. The organization name remains on `organizations` and is updated transactionally with the settings row.

## Portal link administration

Owner and tenant portal users are linked explicitly to business-domain records through `owner_user_links` and `tenant_user_links`. Changing a member away from an owner or tenant role removes stale portal links automatically.

## Reporting

Reporting is read-only and derived from source-of-truth domain tables. The dashboard currently includes:

- properties and units
- occupied units and vacancy rate
- open and emergency maintenance
- leases expiring within 30 and 90 days
- outstanding rent by currency
- overdue rent by currency
- current-month collections by currency

No KPI balance is persisted as a mutable reporting field.

## Audit trail

`GET /api/v1/audit-events` exposes organization-scoped audit history with an upper bound of 500 events per request. The manager UI defaults to the latest 100 events.
