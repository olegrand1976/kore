CREATE TABLE IF NOT EXISTS org.tenant_request_settings (
    tenant_id UUID PRIMARY KEY REFERENCES org.tenants(id) ON DELETE CASCADE,
    channels_enabled JSONB NOT NULL DEFAULT '{"tma":true,"support":true,"maintenance":true}',
    guides_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO org.tenant_request_settings (tenant_id)
SELECT id FROM org.tenants
ON CONFLICT (tenant_id) DO NOTHING;
