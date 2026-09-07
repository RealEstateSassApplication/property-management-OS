CREATE TABLE rent_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    obligation_id UUID NOT NULL,
    adjustment_type TEXT NOT NULL CHECK (adjustment_type IN ('charge', 'late_fee', 'credit', 'writeoff')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    reason TEXT NOT NULL,
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (obligation_id, organization_id)
        REFERENCES rent_obligations(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_rent_adjustments_obligation
    ON rent_adjustments (organization_id, obligation_id, created_at);

CREATE TABLE payment_reversals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payment_id UUID NOT NULL,
    reason TEXT NOT NULL,
    reversed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reversed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, payment_id),
    FOREIGN KEY (payment_id, organization_id)
        REFERENCES payments(id, organization_id) ON DELETE RESTRICT
);

CREATE TABLE security_deposit_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    lease_id UUID NOT NULL,
    required_amount_minor BIGINT NOT NULL CHECK (required_amount_minor >= 0),
    currency CHAR(3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, lease_id),
    UNIQUE (id, organization_id),
    FOREIGN KEY (lease_id, organization_id)
        REFERENCES leases(id, organization_id) ON DELETE RESTRICT
);

CREATE TABLE security_deposit_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    deposit_account_id UUID NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('received', 'deduction', 'refund', 'adjustment_increase', 'adjustment_decrease')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    occurred_on DATE NOT NULL,
    note TEXT NOT NULL,
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (deposit_account_id, organization_id)
        REFERENCES security_deposit_accounts(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_security_deposit_transactions_account
    ON security_deposit_transactions (organization_id, deposit_account_id, occurred_on, created_at);

CREATE TABLE property_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    property_id UUID NOT NULL,
    vendor_id UUID,
    work_order_id UUID,
    category TEXT NOT NULL CHECK (category IN ('maintenance', 'utility', 'tax', 'insurance', 'management', 'cleaning', 'security', 'other')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency CHAR(3) NOT NULL,
    incurred_on DATE NOT NULL,
    note TEXT NOT NULL,
    reference_code TEXT,
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (id, organization_id),
    FOREIGN KEY (property_id, organization_id)
        REFERENCES properties(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (vendor_id, organization_id)
        REFERENCES vendors(id, organization_id) ON DELETE RESTRICT,
    FOREIGN KEY (work_order_id, organization_id)
        REFERENCES work_orders(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_property_expenses_property_date
    ON property_expenses (organization_id, property_id, incurred_on DESC);
CREATE UNIQUE INDEX idx_property_expenses_reference
    ON property_expenses (organization_id, reference_code)
    WHERE reference_code IS NOT NULL;

CREATE TABLE property_expense_reversals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    expense_id UUID NOT NULL,
    reason TEXT NOT NULL,
    reversed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reversed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, expense_id),
    FOREIGN KEY (expense_id, organization_id)
        REFERENCES property_expenses(id, organization_id) ON DELETE RESTRICT
);
