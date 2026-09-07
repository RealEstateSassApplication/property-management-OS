INSERT INTO payments (
    id, organization_id, tenant_id, amount_minor, currency, received_at, method, reference_code, status
) VALUES (
    '34343434-3434-3434-3434-343434343434',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666666',
    7500000, 'LKR', '2026-09-07', 'online', 'DEV-GATEWAY-001', 'posted'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO payment_provider_events (
    id, organization_id, provider, event_id, event_type, payload_sha256,
    status, payment_id, received_at, processed_at
) VALUES (
    '35353535-3535-3535-3535-353535353535',
    '11111111-1111-1111-1111-111111111111',
    'generic_hmac', 'dev-event-001', 'payment.paid',
    repeat('a', 64), 'processed',
    '34343434-3434-3434-3434-343434343434',
    '2026-09-07T10:00:00Z', '2026-09-07T10:00:01Z'
)
ON CONFLICT (provider, event_id) DO NOTHING;
