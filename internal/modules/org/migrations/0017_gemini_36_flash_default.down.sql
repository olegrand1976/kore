ALTER TABLE org.platform_settings
    ALTER COLUMN gemini_model SET DEFAULT 'gemini-3.5-flash';

-- Inverse uniquement les lignes encore non personnalisées (miroir du .up)
UPDATE org.platform_settings
SET gemini_model = 'gemini-3.5-flash',
    updated_at = NOW()
WHERE gemini_model = 'gemini-3.6-flash'
  AND updated_by IS NULL;
