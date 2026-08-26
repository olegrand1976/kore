DROP INDEX IF EXISTS tma.idx_tma_demands_tenant_ticket_number;

ALTER TABLE tma.demands
    DROP COLUMN IF EXISTS ticket_number;

DROP TABLE IF EXISTS tma.tenant_ticket_counters;
