ALTER TABLE org.applications
    ADD COLUMN IF NOT EXISTS methodology_profile TEXT NOT NULL DEFAULT 'psa';

ALTER TABLE org.applications
    DROP CONSTRAINT IF EXISTS org_applications_methodology_profile_check;

ALTER TABLE org.applications
    ADD CONSTRAINT org_applications_methodology_profile_check
    CHECK (methodology_profile IN ('psa', 'agile_scrum', 'agile_kanban'));

CREATE INDEX IF NOT EXISTS idx_org_applications_methodology
    ON org.applications (tenant_id, methodology_profile);
