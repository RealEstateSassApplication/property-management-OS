CREATE TABLE organization_settings (
    organization_id UUID PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    timezone TEXT NOT NULL DEFAULT 'Asia/Colombo',
    default_currency CHAR(3) NOT NULL DEFAULT 'LKR',
    country_code CHAR(2) NOT NULL DEFAULT 'LK',
    billing_email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO organization_settings (organization_id)
SELECT id FROM organizations
ON CONFLICT (organization_id) DO NOTHING;
