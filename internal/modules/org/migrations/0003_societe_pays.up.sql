ALTER TABLE org.societes ADD COLUMN IF NOT EXISTS pays TEXT NOT NULL DEFAULT 'FR';

UPDATE org.societes SET pays = 'FR' WHERE pays IS NULL OR pays = '';
