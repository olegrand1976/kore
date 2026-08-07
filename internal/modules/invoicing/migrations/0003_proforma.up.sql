ALTER TABLE invoicing.invoices
    ADD COLUMN IF NOT EXISTS proforma_token_hash TEXT,
    ADD COLUMN IF NOT EXISTS proforma_recipient_email TEXT,
    ADD COLUMN IF NOT EXISTS proforma_sent_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS proforma_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS proforma_validated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS proforma_rejected_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS proforma_client_comment TEXT,
    ADD COLUMN IF NOT EXISTS invoice_sent_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoicing_invoices_proforma_token
    ON invoicing.invoices (proforma_token_hash)
    WHERE proforma_token_hash IS NOT NULL AND proforma_token_hash <> '';
