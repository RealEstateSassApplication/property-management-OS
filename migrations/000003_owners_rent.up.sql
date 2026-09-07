ALTER TABLE properties
    ADD CONSTRAINT properties_id_organization_unique UNIQUE (id, organization_id);

ALTER TABLE leases
    ADD CONSTRAINT leases_id_organization_unique UNIQUE (id, organization_id);

CREATE TABLE owners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    legal_name TEXT NOT NULL,
    owner_type TEXT NOT NULL DEFAULT 'individual' CHECK (owner_type IN ('individual', 'company')),
    email TEXT,
    phone TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id)
);

CREATE INDEX idx_owners_org_status ON owners (organization_id, status);
CREATE UNIQUE INDEX idx_owners_org_email_unique
    ON owners (organization_id, lower(email))
    WHERE email IS NOT NULL;

CREATE TABLE ownership_interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    property_id UUID NOT NULL,
    owner_id UUID NOT NULL,
    ownership_bps INTEGER NOT NULL CHECK (ownership_bps BETWEEN 1 AND 10000),
    effective_from DATE NOT NULL,
    effective_to DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (effective_to IS NULL OR effective_to >= effective_from),
    UNIQUE (id, organization_id),
    FOREIGN KEY (property_id, organization_id)
        REFERENCES properties(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (owner_id, organization_id)
        REFERENCES owners(id, organization_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_ownership_current_owner_property
    ON ownership_interests (organization_id, property_id, owner_id)
    WHERE effective_to IS NULL;
CREATE INDEX idx_ownership_property ON ownership_interests (organization_id, property_id);
CREATE INDEX idx_ownership_owner ON ownership_interests (organization_id, owner_id);

CREATE TABLE rent_obligations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    lease_id UUID NOT NULL,
    period_start DATE NOT NULL,
    due_date DATE NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency CHAR(3) NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'void')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (date_trunc('month', period_start)::date = period_start),
    UNIQUE (organization_id, lease_id, period_start),
    UNIQUE (id, organization_id),
    FOREIGN KEY (lease_id, organization_id)
        REFERENCES leases(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_rent_obligations_org_due ON rent_obligations (organization_id, due_date);
CREATE INDEX idx_rent_obligations_lease ON rent_obligations (organization_id, lease_id, period_start DESC);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency CHAR(3) NOT NULL,
    received_at DATE NOT NULL,
    method TEXT NOT NULL CHECK (method IN ('cash', 'bank_transfer', 'card', 'online', 'other')),
    reference_code TEXT,
    status TEXT NOT NULL DEFAULT 'posted' CHECK (status IN ('posted', 'void')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (tenant_id, organization_id)
        REFERENCES tenants(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_payments_org_received ON payments (organization_id, received_at DESC);
CREATE INDEX idx_payments_tenant ON payments (organization_id, tenant_id, received_at DESC);
CREATE UNIQUE INDEX idx_payments_org_reference_unique
    ON payments (organization_id, reference_code)
    WHERE reference_code IS NOT NULL;

CREATE TABLE payment_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payment_id UUID NOT NULL,
    obligation_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (payment_id, organization_id)
        REFERENCES payments(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (obligation_id, organization_id)
        REFERENCES rent_obligations(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_payment_allocations_payment ON payment_allocations (organization_id, payment_id);
CREATE INDEX idx_payment_allocations_obligation ON payment_allocations (organization_id, obligation_id);
