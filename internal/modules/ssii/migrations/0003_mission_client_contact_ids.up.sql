ALTER TABLE ssii.missions
    ADD COLUMN IF NOT EXISTS client_contact_ids UUID[] NOT NULL DEFAULT '{}';
