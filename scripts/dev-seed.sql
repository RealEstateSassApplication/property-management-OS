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
