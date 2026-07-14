DROP TABLE IF EXISTS org.user_totp_backup_codes;

ALTER TABLE org.users
    DROP COLUMN IF EXISTS totp_enabled_at,
    DROP COLUMN IF EXISTS totp_secret_encrypted,
    DROP COLUMN IF EXISTS totp_enrollment_required,
    DROP COLUMN IF EXISTS totp_enabled;

ALTER TABLE org.societes
    DROP COLUMN IF EXISTS totp_user_configurable,
    DROP COLUMN IF EXISTS totp_default_enabled;
