CREATE TABLE payment_provider_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    event_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload_sha256 CHAR(64) NOT NULL,
    status TEXT NOT NULL DEFAULT 'received' CHECK (status IN ('received', 'processed', 'rejected')),
    payment_id UUID,
    error_message TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    UNIQUE (provider, event_id),
    FOREIGN KEY (payment_id, organization_id)
        REFERENCES payments(id, organization_id) ON DELETE RESTRICT
);

CREATE INDEX idx_payment_provider_events_org_received
    ON payment_provider_events (organization_id, received_at DESC);
CREATE INDEX idx_payment_provider_events_status
    ON payment_provider_events (status, received_at);
