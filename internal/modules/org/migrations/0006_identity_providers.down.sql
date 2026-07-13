DROP TABLE IF EXISTS org.user_identities;
DROP TABLE IF EXISTS org.identity_providers;
ALTER TABLE org.users DROP COLUMN IF EXISTS email;
