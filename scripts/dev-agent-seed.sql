INSERT INTO users (id, email, display_name, status)
VALUES (
    '17171717-1717-1717-1717-171717171717',
    'property-os-agent@example.invalid',
    'Property OS MCP Agent',
    'active'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO organization_memberships (organization_id, user_id, role)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '17171717-1717-1717-1717-171717171717',
    'agent'
)
ON CONFLICT (organization_id, user_id) DO UPDATE SET role=EXCLUDED.role;

-- Open maintenance fixture used to exercise MCP proposal -> human approval locally.
INSERT INTO maintenance_requests (
    id, organization_id, property_id, unit_id, tenant_id, reported_by_user_id,
    title, description, category, priority, status
) VALUES (
    '18181818-1818-1818-1818-181818181818',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444',
    '66666666-6666-6666-6666-666666666666',
    '22222222-2222-2222-2222-222222222222',
    'Bedroom air conditioner not cooling',
    'The unit runs but does not cool the room; technician inspection and repair are required.',
    'hvac', 'normal', 'triaged'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO work_orders (
    id, organization_id, maintenance_request_id, vendor_id, summary, status,
    scheduled_for
) VALUES (
    '19191919-1919-1919-1919-191919191919',
    '11111111-1111-1111-1111-111111111111',
    '18181818-1818-1818-1818-181818181818',
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    'Inspect AC cooling failure and replace failed components if required',
    'planned', '2026-09-09T04:30:00Z'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO maintenance_quotes (
    id, organization_id, work_order_id, vendor_id, amount_minor, currency,
    scope_summary, status, submitted_at
) VALUES (
    '20202020-2020-2020-2020-202020202020',
    '11111111-1111-1111-1111-111111111111',
    '19191919-1919-1919-1919-191919191919',
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    2750000, 'LKR',
    'Diagnose cooling fault, replace capacitor or equivalent failed component, service unit and verify outlet temperature.',
    'submitted', '2026-09-07T10:45:00Z'
)
ON CONFLICT (id) DO NOTHING;
