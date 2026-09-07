ALTER TABLE organization_memberships DROP CONSTRAINT organization_memberships_role_check;
ALTER TABLE organization_memberships
    ADD CONSTRAINT organization_memberships_role_check
    CHECK (role IN ('admin', 'manager', 'owner', 'tenant', 'accountant', 'maintenance', 'viewer', 'agent'));

CREATE TABLE owner_user_links (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, owner_id, user_id),
    FOREIGN KEY (owner_id, organization_id)
        REFERENCES owners(id, organization_id) ON DELETE CASCADE
);

CREATE INDEX idx_owner_user_links_user
    ON owner_user_links (organization_id, user_id, owner_id);

CREATE TABLE tenant_user_links (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, tenant_id, user_id),
    FOREIGN KEY (tenant_id, organization_id)
        REFERENCES tenants(id, organization_id) ON DELETE CASCADE
);

CREATE INDEX idx_tenant_user_links_user
    ON tenant_user_links (organization_id, user_id, tenant_id);
