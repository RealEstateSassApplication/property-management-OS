# Resource-scoped portals

Owner and tenant portals are separate authorization surfaces from the property-manager application.

```text
OIDC / development identity
        |
        v
organization_memberships
        |
        +--> owner  -> portal:owner:view
        |
        +--> tenant -> portal:tenant:view
                      portal:tenant:create_maintenance_request
        |
        v
owner_user_links / tenant_user_links
        |
        v
scoped domain queries
```

An owner or tenant membership does not grant generalized organization read access. The normal portfolio, owners, rent, people, maintenance and document endpoints remain unavailable to these roles.

## Owner portal

`GET /api/v1/portal/owner` starts from `owner_user_links` for the authenticated user and returns only:

- linked active owner records
- properties covered by current effective ownership interests for those owner records
- unit/occupancy counts for those properties
- open maintenance counts for those properties
- outstanding rent obligations tied to those properties, grouped by currency

The response never accepts an owner ID supplied by the browser.

## Tenant portal

`GET /api/v1/portal/tenant` starts from `tenant_user_links` and returns only:

- linked tenant profiles
- tenancies involving those tenants
- the units/properties for those tenancies
- the most relevant lease for each tenancy
- rent obligations derived through those leases
- maintenance requests whose tenant ID is one of the linked tenant records

## Tenant maintenance intake

`POST /api/v1/portal/tenant/maintenance-requests` accepts:

```json
{
  "tenancyId": "...",
  "title": "AC not cooling",
  "description": "Warm air is coming from the vents",
  "category": "hvac",
  "priority": "high"
}
```

It deliberately does not accept `propertyId`, `unitId` or `tenantId`.

The backend looks up an active tenancy through the authenticated user's `tenant_user_links`, derives the actual property/unit/tenant IDs, and then invokes the existing maintenance service. This gives two independent safeguards:

1. portal-level resource scoping
2. maintenance-domain occupancy validation

## Development fixtures

`make seed` creates:

- owner portal user `18181818-1818-1818-1818-181818181818` linked to owner `99999999-9999-9999-9999-999999999999`
- tenant portal user `19191919-1919-1919-1919-191919191919` linked to tenant `66666666-6666-6666-6666-666666666666`

For local Next.js testing, set one of:

```bash
PROPERTY_OS_OWNER_PORTAL_USER_ID=18181818-1818-1818-1818-181818181818
PROPERTY_OS_TENANT_PORTAL_USER_ID=19191919-1919-1919-1919-191919191919
```

Production should use the logged-in user's Bearer access token instead of development IDs.
