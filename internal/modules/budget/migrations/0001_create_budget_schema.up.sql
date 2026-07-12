CREATE SCHEMA IF NOT EXISTS budget;

CREATE TABLE IF NOT EXISTS budget.budgets (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    application_id UUID NOT NULL,
    type TEXT NOT NULL,
    planned_days NUMERIC(12, 2) NOT NULL DEFAULT 0,
    planned_uo NUMERIC(12, 2) NOT NULL DEFAULT 0,
    planned_amount BIGINT NOT NULL DEFAULT 0,
    consumed_days NUMERIC(12, 2) NOT NULL DEFAULT 0,
    consumed_uo NUMERIC(12, 2) NOT NULL DEFAULT 0,
    consumed_amount BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'EUR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget.estimates (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    budget_id UUID NOT NULL REFERENCES budget.budgets(id),
    demand_id UUID NOT NULL,
    effort_uo NUMERIC(12, 2) NOT NULL DEFAULT 0,
    effort_days NUMERIC(12, 2) NOT NULL DEFAULT 0,
    superseded BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget.quotes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    budget_id UUID NOT NULL REFERENCES budget.budgets(id),
    demand_id UUID NOT NULL,
    amount BIGINT NOT NULL DEFAULT 0,
    effort_uo NUMERIC(12, 2) NOT NULL DEFAULT 0,
    effort_days NUMERIC(12, 2) NOT NULL DEFAULT 0,
    supersedes_estimate_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget.consumptions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    budget_id UUID NOT NULL REFERENCES budget.budgets(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    days NUMERIC(12, 2) NOT NULL DEFAULT 0,
    uo NUMERIC(12, 2) NOT NULL DEFAULT 0,
    amount BIGINT NOT NULL DEFAULT 0,
    approved_at TIMESTAMPTZ,
    approved_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, budget_id, period_start, period_end)
);

CREATE INDEX IF NOT EXISTS idx_budget_budgets_app ON budget.budgets(tenant_id, application_id);
CREATE INDEX IF NOT EXISTS idx_budget_budgets_type ON budget.budgets(tenant_id, application_id, type);
