-- Seed definitions are tenant-agnostic placeholders; tenants receive rows on first access via app seed if needed.
-- This migration documents expected codes for dashboards and TMA reporting.

INSERT INTO reporting.report_definitions (id, tenant_id, code, name, config, active, created_at)
SELECT gen_random_uuid(), t.id, 'tma_summary', 'TMA — synthèse', '{"kind":"tma_summary"}'::jsonb, TRUE, NOW()
FROM org.tenants t
ON CONFLICT (tenant_id, code) DO NOTHING;
