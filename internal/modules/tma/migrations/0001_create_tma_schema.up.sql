CREATE SCHEMA IF NOT EXISTS tma;

CREATE TABLE IF NOT EXISTS tma.demands (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    application_id UUID NOT NULL,
    type TEXT NOT NULL DEFAULT 'incident',
    subject TEXT NOT NULL,
    workflow_instance_id UUID,
    author_id UUID NOT NULL,
    assignee_id UUID,
    status TEXT NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT TRUE,
    consumption_active BOOLEAN NOT NULL DEFAULT TRUE,
    requires_chef_gate BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tma.analysis_dossiers (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    demand_id UUID NOT NULL REFERENCES tma.demands(id),
    functional TEXT NOT NULL DEFAULT '',
    technical TEXT NOT NULL DEFAULT '',
    risks TEXT NOT NULL DEFAULT '',
    test_scenario TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tma.releases (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    application_id UUID NOT NULL,
    label TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tma.delivery_codes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    release_id UUID NOT NULL REFERENCES tma.releases(id),
    code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tma_demands_tenant_app_status ON tma.demands(tenant_id, application_id, status);
