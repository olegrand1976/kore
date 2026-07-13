CREATE SCHEMA IF NOT EXISTS integrations;

CREATE TABLE IF NOT EXISTS integrations.connections (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    type TEXT NOT NULL,
    provider TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    credentials_ref TEXT NOT NULL DEFAULT '',
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS integrations.api_keys (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS integrations.webhook_subscriptions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    secret_ref TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS integrations.sync_jobs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    connection_id UUID NOT NULL REFERENCES integrations.connections(id),
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    error_message TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_integrations_connections_tenant ON integrations.connections(tenant_id);
CREATE INDEX IF NOT EXISTS idx_integrations_api_keys_tenant ON integrations.api_keys(tenant_id);
