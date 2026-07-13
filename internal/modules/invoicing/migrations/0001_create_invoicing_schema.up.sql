CREATE SCHEMA IF NOT EXISTS invoicing;

CREATE TABLE IF NOT EXISTS invoicing.invoices (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    client_id UUID NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'virtuelle',
    currency TEXT NOT NULL DEFAULT 'EUR',
    total_amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    pdp_receipt_id TEXT,
    transmitted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoicing.invoice_lines (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    invoice_id UUID NOT NULL REFERENCES invoicing.invoices(id),
    description TEXT NOT NULL,
    quantity NUMERIC(12, 2) NOT NULL DEFAULT 1,
    unit_price BIGINT NOT NULL DEFAULT 0,
    tax_rate NUMERIC(5, 2) NOT NULL DEFAULT 20
);

CREATE TABLE IF NOT EXISTS invoicing.pdp_queue (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    invoice_id UUID NOT NULL REFERENCES invoicing.invoices(id),
    payload JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_retry_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_invoicing_invoices_tenant ON invoicing.invoices(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_invoicing_pdp_queue_status ON invoicing.pdp_queue(tenant_id, status);