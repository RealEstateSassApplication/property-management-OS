INSERT INTO notification_outbox (
    id, organization_id, actor_user_id, topic, channel, recipient, subject, body,
    resource_type, resource_id, idempotency_key, status
) VALUES (
    '16161616-1616-1616-1616-161616161616',
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    'rent.reminder',
    'email',
    'maya@example.com',
    'Rent reminder',
    'Development seed reminder for the September rent obligation.',
    'rent_obligation',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'dev:rent-reminder:bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'pending'
) ON CONFLICT (id) DO NOTHING;
