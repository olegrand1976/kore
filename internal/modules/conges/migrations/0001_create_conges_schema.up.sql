CREATE SCHEMA IF NOT EXISTS conges;

CREATE TABLE IF NOT EXISTS conges.leave_requests (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    type TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    motif TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'en_attente',
    decided_by UUID,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS conges.leave_balances (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    type TEXT NOT NULL,
    acquired NUMERIC(8, 2) NOT NULL DEFAULT 0,
    taken NUMERIC(8, 2) NOT NULL DEFAULT 0,
    remaining NUMERIC(8, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, type)
);

CREATE INDEX IF NOT EXISTS idx_conges_leave_requests_tenant_user ON conges.leave_requests(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_conges_leave_requests_status ON conges.leave_requests(tenant_id, status);
