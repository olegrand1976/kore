DROP INDEX IF EXISTS invoicing.idx_invoicing_invoices_proforma_token;

ALTER TABLE invoicing.invoices
    DROP COLUMN IF EXISTS proforma_token_hash,
    DROP COLUMN IF EXISTS proforma_recipient_email,
    DROP COLUMN IF EXISTS proforma_sent_at,
    DROP COLUMN IF EXISTS proforma_expires_at,
    DROP COLUMN IF EXISTS proforma_validated_at,
    DROP COLUMN IF EXISTS proforma_rejected_at,
    DROP COLUMN IF EXISTS proforma_client_comment,
    DROP COLUMN IF EXISTS invoice_sent_at;
