CREATE SCHEMA IF NOT EXISTS project;

CREATE TABLE IF NOT EXISTS project.epics (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    application_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    priority TEXT NOT NULL DEFAULT 'medium',
    target_sprint_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT project_epics_status_check CHECK (status IN ('draft', 'active', 'done')),
    CONSTRAINT project_epics_priority_check CHECK (priority IN ('low', 'medium', 'high', 'critical'))
);

CREATE TABLE IF NOT EXISTS project.sprints (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    application_id UUID NOT NULL,
    name TEXT NOT NULL,
    goal TEXT NOT NULL DEFAULT '',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'planned',
    capacity_points SMALLINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    CONSTRAINT project_sprints_status_check CHECK (status IN ('planned', 'active', 'closed'))
);

ALTER TABLE project.epics
    ADD CONSTRAINT project_epics_target_sprint_fk
    FOREIGN KEY (target_sprint_id) REFERENCES project.sprints(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS project.kanban_configs (
    application_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    columns JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_epics_tenant_app ON project.epics (tenant_id, application_id);
CREATE INDEX IF NOT EXISTS idx_project_sprints_tenant_app ON project.sprints (tenant_id, application_id);
CREATE INDEX IF NOT EXISTS idx_project_sprints_status ON project.sprints (tenant_id, application_id, status);
