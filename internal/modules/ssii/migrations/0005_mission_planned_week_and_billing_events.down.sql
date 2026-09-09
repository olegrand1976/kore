DROP INDEX IF EXISTS ssii.idx_ssii_mission_billing_events_mission;
DROP TABLE IF EXISTS ssii.mission_billing_events;

ALTER TABLE ssii.missions
    DROP CONSTRAINT IF EXISTS chk_ssii_missions_planned_week_minutes;

ALTER TABLE ssii.missions
    DROP COLUMN IF EXISTS planned_week_minutes;
