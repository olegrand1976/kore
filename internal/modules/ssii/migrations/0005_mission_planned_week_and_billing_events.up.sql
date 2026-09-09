ALTER TABLE ssii.missions
    ADD COLUMN IF NOT EXISTS planned_week_minutes INT NULL;

ALTER TABLE ssii.missions
    DROP CONSTRAINT IF EXISTS chk_ssii_missions_planned_week_minutes;

ALTER TABLE ssii.missions
    ADD CONSTRAINT chk_ssii_missions_planned_week_minutes
    CHECK (planned_week_minutes IS NULL OR planned_week_minutes > 0);

CREATE TABLE IF NOT EXISTS ssii.mission_billing_events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    mission_id UUID NOT NULL REFERENCES ssii.missions(id) ON DELETE CASCADE,
    actor_user_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ssii_mission_billing_events_mission
    ON ssii.mission_billing_events(tenant_id, mission_id, created_at DESC);
