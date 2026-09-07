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

INSERT INTO owners (id, organization_id, legal_name, owner_type, email, phone, status)
VALUES (
    '99999999-9999-9999-9999-999999999999',
    '11111111-1111-1111-1111-111111111111',
    'Avara Capital Holdings', 'company', 'owners@example.com', '+94 11 555 0101', 'active'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO ownership_interests (
    id, organization_id, property_id, owner_id, ownership_bps, effective_from
) VALUES (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    '99999999-9999-9999-9999-999999999999',
    10000, '2026-01-01'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO rent_obligations (
    id, organization_id, lease_id, period_start, due_date, amount_minor, currency, status
)
SELECT
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    l.organization_id,
    l.id,
    '2026-09-01'::date,
    '2026-09-01'::date,
    l.rent_amount_minor,
    l.currency,
    'open'
FROM leases l
WHERE l.id = '88888888-8888-8888-8888-888888888888'
ON CONFLICT (id) DO NOTHING;

INSERT INTO payments (
    id, organization_id, tenant_id, amount_minor, currency, received_at, method, reference_code, status
) VALUES (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666666',
    10000000, 'LKR', '2026-09-07', 'bank_transfer', 'DEV-PAY-001', 'posted'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO payment_allocations (
    id, organization_id, payment_id, obligation_id, amount_minor
) VALUES (
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    '11111111-1111-1111-1111-111111111111',
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    10000000
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO vendors (id, organization_id, name, trade, email, phone, status)
VALUES (
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    '11111111-1111-1111-1111-111111111111',
    'Colombo Rapid Plumbing', 'plumbing', 'dispatch@rapidplumbing.example', '+94 77 555 0202', 'active'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO maintenance_requests (
    id, organization_id, property_id, unit_id, tenant_id, reported_by_user_id,
    title, description, category, priority, status, resolved_at
) VALUES (
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444',
    '66666666-6666-6666-6666-666666666666',
    '22222222-2222-2222-2222-222222222222',
    'Kitchen sink leak', 'Water was leaking below the kitchen sink cabinet.',
    'plumbing', 'high', 'resolved', '2026-09-07T06:30:00Z'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO work_orders (
    id, organization_id, maintenance_request_id, vendor_id, summary, status,
    scheduled_for, started_at, completed_at
) VALUES (
    '12121212-1212-1212-1212-121212121212',
    '11111111-1111-1111-1111-111111111111',
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    'Replace failed sink trap and reseal connection', 'completed',
    '2026-09-07T05:00:00Z', '2026-09-07T05:10:00Z', '2026-09-07T06:20:00Z'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO maintenance_quotes (
    id, organization_id, work_order_id, vendor_id, amount_minor, currency,
    scope_summary, status, submitted_at, reviewed_at, reviewed_by_user_id
) VALUES (
    '13131313-1313-1313-1313-131313131313',
    '11111111-1111-1111-1111-111111111111',
    '12121212-1212-1212-1212-121212121212',
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    1850000, 'LKR', 'Replace sink trap, fittings, sealant and test for leaks.',
    'approved', '2026-09-07T04:30:00Z', '2026-09-07T04:40:00Z',
    '22222222-2222-2222-2222-222222222222'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO maintenance_completion_evidence (
    id, organization_id, work_order_id, evidence_type, note, submitted_by_user_id
) VALUES (
    '14141414-1414-1414-1414-141414141414',
    '11111111-1111-1111-1111-111111111111',
    '12121212-1212-1212-1212-121212121212',
    'note', 'Sink trap replaced and cabinet area remained dry after a 15-minute flow test.',
    '22222222-2222-2222-2222-222222222222'
)
ON CONFLICT (id) DO NOTHING;
