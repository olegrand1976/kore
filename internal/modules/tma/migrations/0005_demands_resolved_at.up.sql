ALTER TABLE tma.demands ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_tma_demands_sprint_resolved
    ON tma.demands (tenant_id, sprint_id, resolved_at);
