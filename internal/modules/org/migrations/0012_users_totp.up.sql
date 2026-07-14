ALTER TABLE org.societes
    ADD COLUMN IF NOT EXISTS totp_default_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS totp_user_configurable BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE org.users
    ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS totp_enrollment_required BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS totp_secret_encrypted TEXT,
    ADD COLUMN IF NOT EXISTS totp_enabled_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS org.user_totp_backup_codes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    user_id UUID NOT NULL REFERENCES org.users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_totp_backup_codes_user
    ON org.user_totp_backup_codes (user_id)
    WHERE used_at IS NULL;
