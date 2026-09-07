ALTER TABLE organization_memberships DROP CONSTRAINT organization_memberships_role_check;
ALTER TABLE organization_memberships
    ADD CONSTRAINT organization_memberships_role_check
    CHECK (role IN ('admin', 'manager', 'owner', 'accountant', 'maintenance', 'viewer', 'agent'));

CREATE TABLE agent_action_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    proposed_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reviewed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action_type TEXT NOT NULL CHECK (action_type IN ('maintenance.quote.approve')),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('maintenance_quote')),
    resource_id UUID NOT NULL,
    risk_level TEXT NOT NULL DEFAULT 'high' CHECK (risk_level IN ('low', 'medium', 'high')),
    title TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'proposed' CHECK (status IN ('proposed', 'approved', 'rejected', 'executed', 'failed')),
    decision_reason TEXT,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_error TEXT,
    reviewed_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agent_action_requests_org_created
    ON agent_action_requests (organization_id, created_at DESC);
CREATE INDEX idx_agent_action_requests_org_status
    ON agent_action_requests (organization_id, status, created_at DESC);
CREATE UNIQUE INDEX uq_agent_action_active_resource
    ON agent_action_requests (organization_id, action_type, resource_id)
    WHERE status IN ('proposed', 'approved', 'executed');
