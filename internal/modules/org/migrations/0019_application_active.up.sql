ALTER TABLE org.applications
    ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_org_applications_tenant_active
    ON org.applications (tenant_id, active);
