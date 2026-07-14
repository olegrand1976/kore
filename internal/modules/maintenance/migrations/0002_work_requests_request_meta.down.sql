ALTER TABLE maintenance.work_requests
    DROP COLUMN IF EXISTS due_at,
    DROP COLUMN IF EXISTS priority,
    DROP COLUMN IF EXISTS description;
