ALTER TABLE org.societes
  ADD COLUMN IF NOT EXISTS task_types_enabled JSONB NOT NULL DEFAULT '[]';
