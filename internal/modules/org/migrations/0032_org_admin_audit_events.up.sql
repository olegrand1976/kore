CREATE TABLE IF NOT EXISTS org.admin_audit_events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id) ON DELETE CASCADE,
    actor_user_id UUID NOT NULL,
    action TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_org_admin_audit_tenant_created
    ON org.admin_audit_events(tenant_id, created_at DESC);
