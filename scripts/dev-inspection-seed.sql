INSERT INTO inspections (
    id, organization_id, tenancy_id, property_id, unit_id,
    inspection_type, status, scheduled_for, summary,
    completed_at, completed_by_user_id, acknowledged_at, acknowledged_by_user_id,
    created_by_user_id
) VALUES (
    '30303030-3030-3030-3030-303030303030',
    '11111111-1111-1111-1111-111111111111',
    '77777777-7777-7777-7777-777777777777',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444',
    'periodic', 'acknowledged', '2026-09-06T09:00:00Z',
    'Quarterly condition review completed with no critical defects.',
    '2026-09-06T09:35:00Z', '22222222-2222-2222-2222-222222222222',
    '2026-09-06T10:00:00Z', '22222222-2222-2222-2222-222222222222',
    '22222222-2222-2222-2222-222222222222'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO inspection_items (
    id, organization_id, inspection_id, area, item_name, condition, notes
) VALUES
(
    '31313131-3131-3131-3131-313131313131',
    '11111111-1111-1111-1111-111111111111',
    '30303030-3030-3030-3030-303030303030',
    'Kitchen', 'Sink and plumbing', 'good', 'No active leaks; cabinet base dry.'
),
(
    '32323232-3232-3232-3232-323232323232',
    '11111111-1111-1111-1111-111111111111',
    '30303030-3030-3030-3030-303030303030',
    'Living room', 'Walls and ceiling', 'fair', 'Minor paint wear near balcony door; no moisture marks.'
)
ON CONFLICT (id) DO NOTHING;
