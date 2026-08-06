ALTER TABLE invoicing.invoices
    ADD COLUMN IF NOT EXISTS source_timesheet_id UUID;

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoicing_invoices_source_timesheet
    ON invoicing.invoices (tenant_id, source_timesheet_id)
    WHERE source_timesheet_id IS NOT NULL;
