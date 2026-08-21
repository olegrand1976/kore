ALTER TABLE tma.demands ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_tma_demands_tenant_active
    ON tma.demands (tenant_id)
    WHERE deleted_at IS NULL;
