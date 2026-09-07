ALTER TABLE units
    ADD CONSTRAINT units_id_organization_unique UNIQUE (id, organization_id);

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    legal_name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    status TEXT NOT NULL DEFAULT 'prospect' CHECK (status IN ('prospect', 'active', 'former', 'blocked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id)
);

CREATE INDEX idx_tenants_organization_status ON tenants (organization_id, status);
CREATE UNIQUE INDEX idx_tenants_org_email_unique
    ON tenants (organization_id, lower(email))
    WHERE email IS NOT NULL;

CREATE TABLE tenancies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    unit_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    status TEXT NOT NULL DEFAULT 'upcoming' CHECK (status IN ('upcoming', 'active', 'ended', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date IS NULL OR end_date >= start_date),
    UNIQUE (id, organization_id),
    FOREIGN KEY (unit_id, organization_id)
        REFERENCES units(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_tenancies_org_status ON tenancies (organization_id, status);
CREATE INDEX idx_tenancies_unit ON tenancies (organization_id, unit_id);
CREATE UNIQUE INDEX idx_tenancies_one_active_per_unit
    ON tenancies (unit_id)
    WHERE status = 'active';

CREATE TABLE tenancy_tenants (
    organization_id UUID NOT NULL,
    tenancy_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('primary', 'occupant')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenancy_id, tenant_id),
    FOREIGN KEY (tenancy_id, organization_id)
        REFERENCES tenancies(id, organization_id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id, organization_id)
        REFERENCES tenants(id, organization_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_tenancy_tenants_one_primary
    ON tenancy_tenants (tenancy_id)
    WHERE role = 'primary';
CREATE INDEX idx_tenancy_tenants_tenant ON tenancy_tenants (organization_id, tenant_id);

CREATE TABLE leases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenancy_id UUID NOT NULL,
    reference_code TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    rent_amount_minor BIGINT NOT NULL CHECK (rent_amount_minor > 0),
    deposit_amount_minor BIGINT NOT NULL DEFAULT 0 CHECK (deposit_amount_minor >= 0),
    currency CHAR(3) NOT NULL,
    due_day SMALLINT NOT NULL CHECK (due_day BETWEEN 1 AND 31),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'expired', 'terminated', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date >= start_date),
    UNIQUE (organization_id, reference_code),
    FOREIGN KEY (tenancy_id, organization_id)
        REFERENCES tenancies(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_leases_org_status ON leases (organization_id, status);
CREATE INDEX idx_leases_tenancy ON leases (organization_id, tenancy_id);
CREATE UNIQUE INDEX idx_leases_one_active_per_tenancy
    ON leases (tenancy_id)
    WHERE status = 'active';
