CREATE SCHEMA IF NOT EXISTS notifications;

CREATE TABLE IF NOT EXISTS notifications.rules (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    code TEXT NOT NULL,
    trigger TEXT NOT NULL,
    frequency TEXT NOT NULL,
    recipient_policy JSONB NOT NULL DEFAULT '{}',
    template TEXT NOT NULL DEFAULT '',
    attach_pdf BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code),
    UNIQUE (tenant_id, trigger)
);

CREATE TABLE IF NOT EXISTS notifications.messages (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    rule_code TEXT,
    recipients JSONB NOT NULL DEFAULT '[]',
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_rules_tenant ON notifications.rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_messages_tenant ON notifications.messages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_messages_status ON notifications.messages(tenant_id, status);
