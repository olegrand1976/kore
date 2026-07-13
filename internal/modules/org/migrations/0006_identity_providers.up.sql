ALTER TABLE org.users ADD COLUMN IF NOT EXISTS email TEXT;

CREATE TABLE IF NOT EXISTS org.identity_providers (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    name TEXT NOT NULL,
    issuer TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL DEFAULT '',
    jwks_uri TEXT NOT NULL DEFAULT '',
    scopes TEXT NOT NULL DEFAULT 'openid profile email',
    default_profile TEXT NOT NULL DEFAULT 'Collaborateur',
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id)
);

CREATE TABLE IF NOT EXISTS org.user_identities (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    user_id UUID NOT NULL REFERENCES org.users(id),
    idp_id UUID NOT NULL REFERENCES org.identity_providers(id) ON DELETE CASCADE,
    subject TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, idp_id, subject),
    UNIQUE (tenant_id, user_id, idp_id)
);

CREATE INDEX IF NOT EXISTS idx_user_identities_email ON org.user_identities (tenant_id, lower(email));
