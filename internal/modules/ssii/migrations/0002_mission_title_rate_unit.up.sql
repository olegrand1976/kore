ALTER TABLE ssii.missions
    ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS rate_unit TEXT NOT NULL DEFAULT 'tjm';

UPDATE ssii.missions SET rate_unit = 'tjm' WHERE rate_unit IS NULL OR rate_unit = '';
