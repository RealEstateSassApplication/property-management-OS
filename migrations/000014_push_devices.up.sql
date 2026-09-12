CREATE TABLE push_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    expo_push_token TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('ios', 'android')),
    device_name TEXT,
    app_version TEXT,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_push_device_membership
        FOREIGN KEY (organization_id, user_id)
        REFERENCES organization_memberships (organization_id, user_id)
        ON DELETE CASCADE,
    UNIQUE (organization_id, expo_push_token)
);

CREATE INDEX idx_push_devices_user
    ON push_devices (organization_id, user_id, last_seen_at DESC);
