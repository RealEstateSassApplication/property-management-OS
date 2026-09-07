CREATE TABLE inspections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenancy_id UUID NOT NULL,
    property_id UUID NOT NULL,
    unit_id UUID NOT NULL,
    inspection_type TEXT NOT NULL CHECK (inspection_type IN ('move_in', 'move_out', 'periodic')),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'in_progress', 'completed', 'acknowledged', 'cancelled')),
    scheduled_for TIMESTAMPTZ,
    summary TEXT,
    completed_at TIMESTAMPTZ,
    completed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (tenancy_id, organization_id)
        REFERENCES tenancies(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (property_id, organization_id)
        REFERENCES properties(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (unit_id, organization_id)
        REFERENCES units(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_inspections_org_status
    ON inspections (organization_id, status, scheduled_for, created_at DESC);
CREATE INDEX idx_inspections_tenancy
    ON inspections (organization_id, tenancy_id, created_at DESC);

CREATE TABLE inspection_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    inspection_id UUID NOT NULL,
    area TEXT NOT NULL,
    item_name TEXT NOT NULL,
    condition TEXT NOT NULL CHECK (condition IN ('good', 'fair', 'poor', 'damaged', 'not_applicable')),
    notes TEXT,
    evidence_document_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (inspection_id, organization_id)
        REFERENCES inspections(id, organization_id) ON DELETE CASCADE,
    FOREIGN KEY (evidence_document_id, organization_id)
        REFERENCES documents(id, organization_id) ON DELETE SET NULL
);

CREATE INDEX idx_inspection_items_inspection
    ON inspection_items (organization_id, inspection_id, area, item_name);
