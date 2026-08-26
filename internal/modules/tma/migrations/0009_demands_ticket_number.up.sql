CREATE TABLE IF NOT EXISTS tma.tenant_ticket_counters (
    tenant_id UUID PRIMARY KEY,
    last_number INT NOT NULL DEFAULT 0
);

ALTER TABLE tma.demands
    ADD COLUMN IF NOT EXISTS ticket_number INT;

-- Backfill existing rows per tenant (stable order: created_at, id).
WITH ordered AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY tenant_id ORDER BY created_at ASC, id ASC) AS rn
    FROM tma.demands
    WHERE ticket_number IS NULL
)
UPDATE tma.demands d
SET ticket_number = o.rn
FROM ordered o
WHERE d.id = o.id;

INSERT INTO tma.tenant_ticket_counters (tenant_id, last_number)
SELECT tenant_id, COALESCE(MAX(ticket_number), 0)
FROM tma.demands
GROUP BY tenant_id
ON CONFLICT (tenant_id) DO UPDATE
SET last_number = GREATEST(tma.tenant_ticket_counters.last_number, EXCLUDED.last_number);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tma_demands_tenant_ticket_number
    ON tma.demands (tenant_id, ticket_number)
    WHERE ticket_number IS NOT NULL;
