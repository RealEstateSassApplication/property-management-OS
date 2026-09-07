INSERT INTO users (id, email, display_name, status)
VALUES
    ('18181818-1818-1818-1818-181818181818', 'owner-portal@property-os.local', 'Owner Portal User', 'active'),
    ('19191919-1919-1919-1919-191919191919', 'tenant-portal@property-os.local', 'Maya Silva Portal', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO organization_memberships (organization_id, user_id, role)
VALUES
    ('11111111-1111-1111-1111-111111111111', '18181818-1818-1818-1818-181818181818', 'owner'),
    ('11111111-1111-1111-1111-111111111111', '19191919-1919-1919-1919-191919191919', 'tenant')
ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role;

INSERT INTO owner_user_links (organization_id, owner_id, user_id)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '99999999-9999-9999-9999-999999999999',
    '18181818-1818-1818-1818-181818181818'
)
ON CONFLICT DO NOTHING;

INSERT INTO tenant_user_links (organization_id, tenant_id, user_id)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666666',
    '19191919-1919-1919-1919-191919191919'
)
ON CONFLICT DO NOTHING;
