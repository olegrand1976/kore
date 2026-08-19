DROP INDEX IF EXISTS tma.idx_tma_demands_backlog;
DROP INDEX IF EXISTS tma.idx_tma_demands_sprint;
DROP INDEX IF EXISTS tma.idx_tma_demands_epic;

ALTER TABLE tma.demands
    DROP COLUMN IF EXISTS backlog_rank,
    DROP COLUMN IF EXISTS story_points,
    DROP COLUMN IF EXISTS sprint_id,
    DROP COLUMN IF EXISTS epic_id;
