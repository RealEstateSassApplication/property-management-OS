INSERT INTO documents (
    id, organization_id, resource_type, resource_id, kind, file_name,
    content_type, size_bytes, storage_key, status, uploaded_by_user_id, verified_at
) VALUES (
    '15151515-1515-1515-1515-151515151515',
    '11111111-1111-1111-1111-111111111111',
    'lease',
    '88888888-8888-8888-8888-888888888888',
    'lease',
    'lease-col-001-a01-2026.pdf',
    'application/pdf',
    245760,
    '11111111-1111-1111-1111-111111111111/documents/2026/09/15151515-1515-1515-1515-151515151515-lease-col-001-a01-2026.pdf',
    'available',
    '22222222-2222-2222-2222-222222222222',
    '2026-09-07T08:00:00Z'
)
ON CONFLICT (id) DO NOTHING;
