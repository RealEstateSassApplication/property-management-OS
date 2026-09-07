DROP TABLE IF EXISTS tenant_user_links;
DROP TABLE IF EXISTS owner_user_links;

ALTER TABLE organization_memberships DROP CONSTRAINT organization_memberships_role_check;
UPDATE organization_memberships SET role='viewer' WHERE role='tenant';
ALTER TABLE organization_memberships
    ADD CONSTRAINT organization_memberships_role_check
    CHECK (role IN ('admin', 'manager', 'owner', 'accountant', 'maintenance', 'viewer', 'agent'));
