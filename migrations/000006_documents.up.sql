CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL CHECK (resource_type IN (
        'organization', 'property', 'unit', 'tenant', 'lease', 'owner',
        'rent_payment', 'maintenance_request', 'work_order', 'vendor'
    )),
    resource_id UUID NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN (
        'lease', 'identity', 'inspection', 'invoice', 'receipt',
        'maintenance', 'statement', 'other'
    )),
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 26214400),
    storage_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'available', 'quarantined', 'deleted')),
    checksum_sha256 TEXT,
    uploaded_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, storage_key)
);

CREATE INDEX idx_documents_org_created ON documents (organization_id, created_at DESC);
CREATE INDEX idx_documents_resource ON documents (organization_id, resource_type, resource_id, created_at DESC);
CREATE INDEX idx_documents_status ON documents (organization_id, status) WHERE status <> 'deleted';
