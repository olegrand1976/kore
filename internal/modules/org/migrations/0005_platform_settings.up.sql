-- Singleton des paramètres plateforme (admin multi-tenant)
CREATE TABLE IF NOT EXISTS org.platform_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    gemini_model TEXT NOT NULL DEFAULT 'gemini-3.5-flash',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID
);

INSERT INTO org.platform_settings (id, gemini_model)
VALUES (1, 'gemini-3.5-flash')
ON CONFLICT (id) DO NOTHING;
