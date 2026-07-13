CREATE TABLE IF NOT EXISTS conges.leave_type_configs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    societe_id UUID NOT NULL REFERENCES org.societes(id),
    code TEXT NOT NULL,
    label TEXT NOT NULL,
    tracks_balance BOOLEAN NOT NULL DEFAULT TRUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, societe_id, code)
);

CREATE INDEX IF NOT EXISTS idx_conges_leave_type_configs_tenant_societe
    ON conges.leave_type_configs(tenant_id, societe_id);
