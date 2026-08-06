DROP INDEX IF EXISTS invoicing.idx_invoicing_invoices_source_timesheet;

ALTER TABLE invoicing.invoices
    DROP COLUMN IF EXISTS source_timesheet_id;
