DROP INDEX IF EXISTS org.idx_org_applications_methodology;

ALTER TABLE org.applications
    DROP CONSTRAINT IF EXISTS org_applications_methodology_profile_check;

ALTER TABLE org.applications
    DROP COLUMN IF EXISTS methodology_profile;
