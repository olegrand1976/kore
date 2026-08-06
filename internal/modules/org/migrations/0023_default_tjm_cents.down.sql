ALTER TABLE org.applications
    DROP COLUMN IF EXISTS default_tjm_cents;

ALTER TABLE org.societes
    DROP COLUMN IF EXISTS default_tjm_cents;
