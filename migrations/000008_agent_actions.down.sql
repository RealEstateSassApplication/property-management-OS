DROP TABLE IF EXISTS agent_action_requests;

ALTER TABLE organization_memberships DROP CONSTRAINT organization_memberships_role_check;
UPDATE organization_memberships SET role='viewer' WHERE role='agent';
ALTER TABLE organization_memberships
    ADD CONSTRAINT organization_memberships_role_check
    CHECK (role IN ('admin', 'manager', 'owner', 'accountant', 'maintenance', 'viewer'));
