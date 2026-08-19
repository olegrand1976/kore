DROP INDEX IF EXISTS tma.idx_tma_demands_sprint_resolved;

ALTER TABLE tma.demands DROP COLUMN IF EXISTS resolved_at;
