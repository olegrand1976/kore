UPDATE project.epics SET priority = 'normal' WHERE priority = 'medium';
UPDATE project.epics SET priority = 'urgent' WHERE priority = 'critical';

ALTER TABLE project.epics DROP CONSTRAINT IF EXISTS project_epics_priority_check;

ALTER TABLE project.epics
    ADD CONSTRAINT project_epics_priority_check
    CHECK (priority IN ('low', 'normal', 'high', 'urgent'));

ALTER TABLE project.epics ALTER COLUMN priority SET DEFAULT 'normal';
