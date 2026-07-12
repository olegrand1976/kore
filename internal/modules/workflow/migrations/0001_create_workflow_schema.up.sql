CREATE SCHEMA IF NOT EXISTS workflow;

CREATE TABLE IF NOT EXISTS workflow.definitions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    code TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS workflow.states (
    id UUID PRIMARY KEY,
    definition_id UUID NOT NULL REFERENCES workflow.definitions(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    label TEXT NOT NULL DEFAULT '',
    is_initial BOOLEAN NOT NULL DEFAULT FALSE,
    is_final BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (definition_id, code)
);

CREATE TABLE IF NOT EXISTS workflow.transitions (
    id UUID PRIMARY KEY,
    definition_id UUID NOT NULL REFERENCES workflow.definitions(id) ON DELETE CASCADE,
    from_state TEXT NOT NULL,
    to_state TEXT NOT NULL,
    action TEXT NOT NULL,
    guard TEXT NOT NULL DEFAULT '',
    doc_trigger JSONB,
    allowed_roles TEXT[] NOT NULL DEFAULT '{}',
    UNIQUE (definition_id, from_state, action)
);

CREATE TABLE IF NOT EXISTS workflow.instances (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    definition_code TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    current_state TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflow_instances_tenant_entity
    ON workflow.instances(tenant_id, entity_id);

CREATE TABLE IF NOT EXISTS workflow.transition_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    instance_id UUID NOT NULL REFERENCES workflow.instances(id) ON DELETE CASCADE,
    from_state TEXT NOT NULL,
    to_state TEXT NOT NULL,
    action TEXT NOT NULL,
    actor_id UUID NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflow_transition_logs_instance
    ON workflow.transition_logs(instance_id, occurred_at);
