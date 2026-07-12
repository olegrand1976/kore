CREATE SCHEMA IF NOT EXISTS billing;

CREATE TABLE IF NOT EXISTS billing.subscriptions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL UNIQUE,
    stripe_customer_id TEXT UNIQUE,
    stripe_subscription_id TEXT UNIQUE,
    status TEXT NOT NULL DEFAULT 'trial',
    seats INT NOT NULL DEFAULT 1 CHECK (seats >= 0),
    current_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing.module_entitlements (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    module_code TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, module_code)
);

CREATE TABLE IF NOT EXISTS billing.webhook_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_tenant ON billing.subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_billing_entitlements_tenant ON billing.module_entitlements(tenant_id);
