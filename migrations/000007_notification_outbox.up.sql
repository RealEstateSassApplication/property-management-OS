CREATE TABLE notification_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    topic TEXT NOT NULL,
    channel TEXT NOT NULL CHECK (channel IN ('email', 'sms', 'whatsapp', 'webhook')),
    recipient TEXT NOT NULL,
    subject TEXT,
    body TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    resource_type TEXT,
    resource_id UUID,
    idempotency_key TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'retry', 'delivered', 'dead')),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts INTEGER NOT NULL DEFAULT 5 CHECK (max_attempts BETWEEN 1 AND 20),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_at TIMESTAMPTZ,
    locked_by TEXT,
    last_error TEXT,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT notification_resource_pair CHECK ((resource_type IS NULL) = (resource_id IS NULL))
);

CREATE UNIQUE INDEX notification_outbox_org_idempotency_unique
    ON notification_outbox (organization_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX notification_outbox_delivery_queue_idx
    ON notification_outbox (status, available_at, created_at)
    WHERE status IN ('pending', 'retry', 'processing');

CREATE INDEX notification_outbox_org_created_idx
    ON notification_outbox (organization_id, created_at DESC);
