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
