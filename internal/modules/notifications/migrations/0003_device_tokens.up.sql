CREATE TABLE IF NOT EXISTS notifications.device_tokens (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, token)
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_tenant_user
    ON notifications.device_tokens (tenant_id, user_id);
