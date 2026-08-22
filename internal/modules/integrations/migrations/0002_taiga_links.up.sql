CREATE TABLE IF NOT EXISTS integrations.external_links (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    provider TEXT NOT NULL DEFAULT 'taiga',
    external_type TEXT NOT NULL,
    external_id TEXT NOT NULL,
    external_project_id INT,
    external_ref INT,
    external_url TEXT NOT NULL DEFAULT '',
    kore_entity_type TEXT NOT NULL DEFAULT 'demand',
    kore_entity_id UUID NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, provider, external_type, external_id)
);

CREATE INDEX IF NOT EXISTS idx_integrations_external_links_kore
    ON integrations.external_links (tenant_id, kore_entity_type, kore_entity_id);

CREATE TABLE IF NOT EXISTS integrations.user_mappings (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    provider TEXT NOT NULL DEFAULT 'taiga',
    external_user_id TEXT NOT NULL,
    external_username TEXT NOT NULL DEFAULT '',
    kore_user_id UUID NOT NULL,
    match_method TEXT NOT NULL DEFAULT 'email',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, provider, external_user_id),
    UNIQUE (tenant_id, provider, kore_user_id)
);

CREATE INDEX IF NOT EXISTS idx_integrations_user_mappings_tenant
    ON integrations.user_mappings (tenant_id, provider);
