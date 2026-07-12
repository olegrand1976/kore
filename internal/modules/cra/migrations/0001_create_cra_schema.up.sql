CREATE SCHEMA IF NOT EXISTS cra;

CREATE TABLE IF NOT EXISTS cra.timesheets (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    month TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Brouillon',
    commercial_info JSONB NOT NULL DEFAULT '{}',
    validated_at TIMESTAMPTZ,
    validated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, month)
);

CREATE TABLE IF NOT EXISTS cra.week_entries (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    timesheet_id UUID NOT NULL REFERENCES cra.timesheets(id) ON DELETE CASCADE,
    week_number INT NOT NULL,
    submitted_at TIMESTAMPTZ,
    UNIQUE (timesheet_id, week_number)
);

CREATE TABLE IF NOT EXISTS cra.time_lines (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    week_entry_id UUID NOT NULL REFERENCES cra.week_entries(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,
    day DATE NOT NULL,
    duration INT NOT NULL DEFAULT 0,
    comment TEXT NOT NULL DEFAULT '',
    origin TEXT NOT NULL DEFAULT 'manual',
    UNIQUE (week_entry_id, source_type, source_id, day)
);

CREATE INDEX IF NOT EXISTS idx_cra_timesheets_tenant_user_month
    ON cra.timesheets(tenant_id, user_id, month);

CREATE INDEX IF NOT EXISTS idx_cra_time_lines_source
    ON cra.time_lines(tenant_id, source_type, source_id);
