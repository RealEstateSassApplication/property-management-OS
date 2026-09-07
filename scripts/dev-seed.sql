INSERT INTO organizations (id, name, slug)
VALUES ('11111111-1111-1111-1111-111111111111', 'Avara Property Operations', 'avara-property-operations')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, email, display_name, status)
VALUES ('22222222-2222-2222-2222-222222222222', 'dev@property-os.local', 'Local Property Manager', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO organization_memberships (organization_id, user_id, role)
VALUES ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'admin')
ON CONFLICT (organization_id, user_id) DO NOTHING;

INSERT INTO properties (
    id, organization_id, reference_code, name, property_type,
    address_line_1, city, region, country_code, status
) VALUES (
    '33333333-3333-3333-3333-333333333333',
    '11111111-1111-1111-1111-111111111111',
    'COL-001', 'Park Residences', 'building',
    '18 Park Road', 'Colombo 05', 'Western', 'LK', 'active'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO units (
    id, organization_id, property_id, reference_code, label,
    bedrooms, bathrooms, floor_area, floor_area_unit, occupancy_status
) VALUES
(
    '44444444-4444-4444-4444-444444444444',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    'A-01', 'Apartment 01', 2, 2, 980, 'sqft', 'occupied'
),
(
    '55555555-5555-5555-5555-555555555555',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    'A-02', 'Apartment 02', 2, 2, 980, 'sqft', 'vacant'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tenants (id, organization_id, legal_name, email, phone, status)
VALUES (
    '66666666-6666-6666-6666-666666666666',
    '11111111-1111-1111-1111-111111111111',
    'Maya Silva', 'maya@example.com', '+94 77 555 0101', 'active'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tenancies (id, organization_id, unit_id, start_date, status)
VALUES (
    '77777777-7777-7777-7777-777777777777',
    '11111111-1111-1111-1111-111111111111',
    '44444444-4444-4444-4444-444444444444',
    '2026-01-01', 'active'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO tenancy_tenants (organization_id, tenancy_id, tenant_id, role)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '77777777-7777-7777-7777-777777777777',
    '66666666-6666-6666-6666-666666666666',
    'primary'
)
ON CONFLICT (tenancy_id, tenant_id) DO NOTHING;

INSERT INTO leases (
    id, organization_id, tenancy_id, reference_code, start_date, end_date,
    rent_amount_minor, deposit_amount_minor, currency, due_day, status
) VALUES (
    '88888888-8888-8888-8888-888888888888',
    '11111111-1111-1111-1111-111111111111',
    '77777777-7777-7777-7777-777777777777',
    'LEASE-COL-001-A01-2026', '2026-01-01', '2026-12-31',
    15000000, 30000000, 'LKR', 1, 'active'
)
ON CONFLICT (id) DO NOTHING;
