-- Nouveau défaut plateforme : Gemini 3.6 Flash (meilleur rapport perf/coût que 3.5 Flash)
ALTER TABLE org.platform_settings
    ALTER COLUMN gemini_model SET DEFAULT 'gemini-3.6-flash';

-- Adopte le nouveau défaut uniquement si l'ancien défaut n'a jamais été personnalisé
UPDATE org.platform_settings
SET gemini_model = 'gemini-3.6-flash',
    updated_at = NOW()
WHERE gemini_model = 'gemini-3.5-flash'
  AND updated_by IS NULL;
