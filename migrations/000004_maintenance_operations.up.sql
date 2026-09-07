CREATE TABLE vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    trade TEXT NOT NULL CHECK (trade IN ('plumbing', 'electrical', 'hvac', 'appliance', 'structural', 'cleaning', 'security', 'general', 'other')),
    email TEXT,
    phone TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id)
);

CREATE INDEX idx_vendors_org_trade_status ON vendors (organization_id, trade, status);
CREATE UNIQUE INDEX idx_vendors_org_email_unique
    ON vendors (organization_id, lower(email))
    WHERE email IS NOT NULL;

CREATE TABLE maintenance_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    property_id UUID NOT NULL,
    unit_id UUID,
    tenant_id UUID,
    reported_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('plumbing', 'electrical', 'hvac', 'appliance', 'structural', 'cleaning', 'security', 'other')),
    priority TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'emergency')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'triaged', 'in_progress', 'resolved', 'cancelled')),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (property_id, organization_id)
        REFERENCES properties(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (unit_id, organization_id)
        REFERENCES units(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (tenant_id, organization_id)
        REFERENCES tenants(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_maintenance_requests_org_status ON maintenance_requests (organization_id, status, priority, created_at DESC);
CREATE INDEX idx_maintenance_requests_property ON maintenance_requests (organization_id, property_id, created_at DESC);
CREATE INDEX idx_maintenance_requests_unit ON maintenance_requests (organization_id, unit_id, created_at DESC) WHERE unit_id IS NOT NULL;

CREATE TABLE work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    maintenance_request_id UUID NOT NULL,
    vendor_id UUID,
    assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    summary TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'planned' CHECK (status IN ('planned', 'assigned', 'in_progress', 'completed', 'cancelled')),
    scheduled_for TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (maintenance_request_id, organization_id)
        REFERENCES maintenance_requests(id, organization_id) ON DELETE CASCADE,
    FOREIGN KEY (vendor_id, organization_id)
        REFERENCES vendors(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_work_orders_org_status ON work_orders (organization_id, status, scheduled_for);
CREATE INDEX idx_work_orders_request ON work_orders (organization_id, maintenance_request_id, created_at DESC);
CREATE INDEX idx_work_orders_vendor ON work_orders (organization_id, vendor_id, status) WHERE vendor_id IS NOT NULL;

CREATE TABLE maintenance_quotes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    work_order_id UUID NOT NULL,
    vendor_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency CHAR(3) NOT NULL,
    scope_summary TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted', 'approved', 'rejected', 'withdrawn')),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (work_order_id, organization_id)
        REFERENCES work_orders(id, organization_id) ON DELETE CASCADE,
    FOREIGN KEY (vendor_id, organization_id)
        REFERENCES vendors(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_maintenance_quotes_work_order ON maintenance_quotes (organization_id, work_order_id, submitted_at DESC);
CREATE UNIQUE INDEX idx_maintenance_quotes_one_approved_per_work_order
    ON maintenance_quotes (work_order_id)
    WHERE status = 'approved';

CREATE TABLE maintenance_completion_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    work_order_id UUID NOT NULL,
    evidence_type TEXT NOT NULL CHECK (evidence_type IN ('note', 'photo', 'invoice', 'receipt', 'other')),
    note TEXT,
    storage_key TEXT,
    submitted_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (note IS NOT NULL OR storage_key IS NOT NULL),
    FOREIGN KEY (work_order_id, organization_id)
        REFERENCES work_orders(id, organization_id) ON DELETE CASCADE
);

CREATE INDEX idx_maintenance_evidence_work_order ON maintenance_completion_evidence (organization_id, work_order_id, created_at DESC);
