ALTER TABLE tma.demands
    DROP COLUMN IF EXISTS due_at,
    DROP COLUMN IF EXISTS priority,
    DROP COLUMN IF EXISTS description;
