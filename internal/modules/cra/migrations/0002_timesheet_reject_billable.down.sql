ALTER TABLE cra.time_lines DROP COLUMN IF EXISTS billable;
ALTER TABLE cra.timesheets
  DROP COLUMN IF EXISTS reject_reason,
  DROP COLUMN IF EXISTS rejected_by,
  DROP COLUMN IF EXISTS rejected_at;
