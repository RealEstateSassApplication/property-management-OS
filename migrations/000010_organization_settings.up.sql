CREATE TABLE organization_settings (
    organization_id UUID PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    timezone TEXT NOT NULL DEFAULT 'Asia/Colombo',
    default_currency CHAR(3) NOT NULL DEFAULT 'LKR',
    country_code CHAR(2) NOT NULL DEFAULT 'LK',
    billing_email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE FUNCTION create_default_organization_settings()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO organization_settings (organization_id)
    VALUES (NEW.id)
    ON CONFLICT (organization_id) DO NOTHING;
    RETURN NEW;
END;
$$;

CREATE TRIGGER organizations_create_default_settings
AFTER INSERT ON organizations
FOR EACH ROW
EXECUTE FUNCTION create_default_organization_settings();

INSERT INTO organization_settings (organization_id)
SELECT id FROM organizations
ON CONFLICT (organization_id) DO NOTHING;
