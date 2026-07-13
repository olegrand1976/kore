CREATE SCHEMA IF NOT EXISTS support;

CREATE TABLE IF NOT EXISTS support.tickets (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    application_id UUID NOT NULL,
    subject TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'open',
    channel TEXT NOT NULL DEFAULT 'web',
    reporter_id UUID,
    assignee_id UUID,
    analysis_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS support.ticket_replies (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    ticket_id UUID NOT NULL REFERENCES support.tickets(id),
    author_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_tenant ON support.tickets(tenant_id, state);
CREATE INDEX IF NOT EXISTS idx_support_replies_ticket ON support.ticket_replies(tenant_id, ticket_id);