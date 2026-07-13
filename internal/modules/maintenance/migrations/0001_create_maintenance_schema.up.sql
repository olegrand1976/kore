CREATE SCHEMA IF NOT EXISTS maintenance;

CREATE TABLE IF NOT EXISTS maintenance.work_requests (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    application_id UUID NOT NULL,
    subject TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'created',
    assignee_id UUID,
    consumption_days NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_maintenance_requests_tenant ON maintenance.work_requests(tenant_id, state);