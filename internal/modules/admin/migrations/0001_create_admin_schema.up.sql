CREATE SCHEMA IF NOT EXISTS admin;

CREATE TABLE IF NOT EXISTS admin.parameter_sets (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    code TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS admin.templates (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    content JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS admin.phone_directory (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID,
    label TEXT NOT NULL,
    phone TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'internal',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_parameter_sets_tenant ON admin.parameter_sets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_admin_templates_tenant ON admin.templates(tenant_id);