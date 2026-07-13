CREATE SCHEMA IF NOT EXISTS ett;

CREATE TABLE IF NOT EXISTS ett.work_time_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    work_date DATE NOT NULL,
    clock_in TIMESTAMPTZ,
    clock_out TIMESTAMPTZ,
    effective_hours NUMERIC(5, 2) NOT NULL DEFAULT 0,
    overtime_hours NUMERIC(5, 2) NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pointed',
    origin TEXT NOT NULL DEFAULT 'web',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, work_date)
);

CREATE TABLE IF NOT EXISTS ett.audit_journal (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    record_id UUID NOT NULL,
    action TEXT NOT NULL,
    actor_id UUID NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ett.country_work_rules (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    country_code TEXT NOT NULL,
    max_daily_hours NUMERIC(5, 2) NOT NULL DEFAULT 10,
    min_rest_hours NUMERIC(5, 2) NOT NULL DEFAULT 11,
    retention_days INT NOT NULL DEFAULT 1825,
    UNIQUE (tenant_id, country_code)
);

CREATE INDEX IF NOT EXISTS idx_ett_records_tenant_user ON ett.work_time_records(tenant_id, user_id, work_date);
CREATE INDEX IF NOT EXISTS idx_ett_audit_record ON ett.audit_journal(tenant_id, record_id);