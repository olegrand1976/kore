CREATE TABLE IF NOT EXISTS org.access_tokens (
    token_hash TEXT PRIMARY KEY,
    tenant_id UUID NOT NULL,
    email TEXT NOT NULL,
    kind TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS access_tokens_tenant_email_kind_idx
    ON org.access_tokens (tenant_id, email, kind);

CREATE INDEX IF NOT EXISTS access_tokens_expires_at_idx
    ON org.access_tokens (expires_at);

