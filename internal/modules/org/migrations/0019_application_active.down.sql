DROP INDEX IF EXISTS org.idx_org_applications_tenant_active;
ALTER TABLE org.applications DROP COLUMN IF EXISTS active;
