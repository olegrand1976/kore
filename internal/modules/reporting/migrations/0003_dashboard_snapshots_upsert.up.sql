-- Deduplicate before unique index (legacy rows had no uniqueness).
DELETE FROM reporting.dashboard_snapshots
WHERE id IN (
    SELECT id FROM (
        SELECT id,
               ROW_NUMBER() OVER (
                   PARTITION BY tenant_id, dashboard_code, period_start, period_end
                   ORDER BY computed_at DESC, id DESC
               ) AS rn
        FROM reporting.dashboard_snapshots
    ) ranked
    WHERE rn > 1
);

-- Unique key for upsert of dashboard snapshots (tenant × code × period).
CREATE UNIQUE INDEX IF NOT EXISTS uq_reporting_dashboard_snapshots_period
    ON reporting.dashboard_snapshots (tenant_id, dashboard_code, period_start, period_end);

CREATE INDEX IF NOT EXISTS idx_reporting_dashboard_snapshots_tenant_code
    ON reporting.dashboard_snapshots (tenant_id, dashboard_code, computed_at DESC);
