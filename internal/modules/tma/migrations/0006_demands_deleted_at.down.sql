DROP INDEX IF EXISTS tma.idx_tma_demands_tenant_active;

ALTER TABLE tma.demands DROP COLUMN IF EXISTS deleted_at;
