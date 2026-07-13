ALTER TABLE org.users
    ADD COLUMN IF NOT EXISTS release_notes_auto_show BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS last_seen_version TEXT;

