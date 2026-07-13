CREATE SCHEMA IF NOT EXISTS ai;

CREATE TABLE ai.ai_capabilities (
    code TEXT PRIMARY KEY,
    risk_class TEXT NOT NULL,
    annex_iii BOOLEAN NOT NULL DEFAULT FALSE,
    art_6_3_assessment TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    wave INT NOT NULL DEFAULT 0
);

CREATE TABLE ai.tenant_ai_settings (
    tenant_id UUID PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    notice_accepted_at TIMESTAMPTZ,
    notice_accepted_by UUID,
    workers_informed_at TIMESTAMPTZ,
    llm_provider TEXT NOT NULL DEFAULT 'stub'
);

CREATE TABLE ai.ai_request_log (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    capability_code TEXT NOT NULL REFERENCES ai.ai_capabilities (code),
    entity_type TEXT,
    entity_id UUID,
    input_hash TEXT NOT NULL,
    output_json JSONB,
    model TEXT NOT NULL,
    explain_context JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ai_request_log_tenant_created_idx ON ai.ai_request_log (tenant_id, created_at DESC);
CREATE INDEX ai_request_log_tenant_capability_idx ON ai.ai_request_log (tenant_id, capability_code);

INSERT INTO ai.ai_capabilities (code, risk_class, annex_iii, wave) VALUES
    ('tma.analysis_draft', 'limited', FALSE, 1),
    ('tma.classify', 'minimal', FALSE, 1),
    ('tma.similar', 'minimal', FALSE, 1),
    ('cra.prefill', 'minimal', FALSE, 2),
    ('cra.anomalies', 'minimal', FALSE, 2),
    ('budget.estimate', 'minimal', FALSE, 2),
    ('budget.demand_suggest', 'minimal', FALSE, 2),
    ('dashboard.briefing', 'limited', FALSE, 2),
    ('conges.manager_assist', 'limited', TRUE, 3),
    ('workflow.explain', 'minimal', FALSE, 3),
    ('publicsite.chatbot', 'limited', FALSE, 3);

INSERT INTO ai.tenant_ai_settings (tenant_id, enabled, notice_accepted_at, workers_informed_at, llm_provider)
VALUES ('00000000-0000-4000-8000-000000000001', TRUE, NOW(), NOW(), 'stub')
ON CONFLICT (tenant_id) DO NOTHING;
