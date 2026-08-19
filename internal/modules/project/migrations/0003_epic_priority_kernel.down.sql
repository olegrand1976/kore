UPDATE project.epics SET priority = 'medium' WHERE priority = 'normal';
UPDATE project.epics SET priority = 'critical' WHERE priority = 'urgent';

ALTER TABLE project.epics DROP CONSTRAINT IF EXISTS project_epics_priority_check;

ALTER TABLE project.epics
    ADD CONSTRAINT project_epics_priority_check
    CHECK (priority IN ('low', 'medium', 'high', 'critical'));

ALTER TABLE project.epics ALTER COLUMN priority SET DEFAULT 'medium';
