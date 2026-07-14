ALTER TABLE support.tickets
    DROP COLUMN IF EXISTS due_at,
    DROP COLUMN IF EXISTS priority;
