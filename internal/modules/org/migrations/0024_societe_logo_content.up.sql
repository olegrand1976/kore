-- Durable tenant logo storage (Cloud Run has no persistent local disk).
ALTER TABLE org.societes
    ADD COLUMN IF NOT EXISTS logo_content BYTEA,
    ADD COLUMN IF NOT EXISTS logo_content_type TEXT;
