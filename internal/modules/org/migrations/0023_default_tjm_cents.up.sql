ALTER TABLE org.societes
    ADD COLUMN IF NOT EXISTS default_tjm_cents BIGINT NOT NULL DEFAULT 0
        CHECK (default_tjm_cents >= 0);

ALTER TABLE org.applications
    ADD COLUMN IF NOT EXISTS default_tjm_cents BIGINT NOT NULL DEFAULT 0
        CHECK (default_tjm_cents >= 0);
