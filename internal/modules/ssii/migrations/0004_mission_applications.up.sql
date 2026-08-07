CREATE TABLE IF NOT EXISTS ssii.mission_applications (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    mission_id UUID NOT NULL REFERENCES ssii.missions(id) ON DELETE CASCADE,
    application_id UUID NOT NULL,
    UNIQUE (mission_id, application_id)
);

CREATE INDEX IF NOT EXISTS idx_ssii_mission_applications_tenant
    ON ssii.mission_applications(tenant_id, mission_id);

CREATE INDEX IF NOT EXISTS idx_ssii_mission_applications_app
    ON ssii.mission_applications(tenant_id, application_id);
