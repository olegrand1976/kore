ALTER TABLE tma.demands
    ADD COLUMN IF NOT EXISTS epic_id UUID REFERENCES project.epics(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS sprint_id UUID REFERENCES project.sprints(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS story_points SMALLINT,
    ADD COLUMN IF NOT EXISTS backlog_rank INTEGER;

CREATE INDEX IF NOT EXISTS idx_tma_demands_epic ON tma.demands (tenant_id, epic_id);
CREATE INDEX IF NOT EXISTS idx_tma_demands_sprint ON tma.demands (tenant_id, sprint_id);
CREATE INDEX IF NOT EXISTS idx_tma_demands_backlog ON tma.demands (tenant_id, application_id, backlog_rank);
