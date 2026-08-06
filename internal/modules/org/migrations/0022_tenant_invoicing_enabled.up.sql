ALTER TABLE org.tenant_request_settings
    ADD COLUMN IF NOT EXISTS invoicing_enabled BOOLEAN NOT NULL DEFAULT FALSE;
