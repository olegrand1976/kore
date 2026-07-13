CREATE SCHEMA IF NOT EXISTS ssii;

CREATE TABLE IF NOT EXISTS ssii.missions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    client_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    start_date DATE NOT NULL,
    end_date DATE,
    tjm_amount BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'EUR',
    technologies TEXT[] NOT NULL DEFAULT '{}',
    client_contact TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ssii.mission_collaborators (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    mission_id UUID NOT NULL REFERENCES ssii.missions(id),
    user_id UUID NOT NULL,
    UNIQUE (mission_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ssii_missions_tenant ON ssii.missions(tenant_id, client_id);
CREATE INDEX IF NOT EXISTS idx_ssii_missions_status ON ssii.missions(tenant_id, status);