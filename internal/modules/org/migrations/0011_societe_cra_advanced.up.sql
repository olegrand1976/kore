ALTER TABLE org.societes
  ADD COLUMN IF NOT EXISTS day_capacity_minutes INT NOT NULL DEFAULT 480
    CHECK (day_capacity_minutes > 0 AND day_capacity_minutes <= 1440),
  ADD COLUMN IF NOT EXISTS cra_mail_auto BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS cra_mail_recipients JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS week_submit_policy TEXT NOT NULL DEFAULT 'warn'
    CHECK (week_submit_policy IN ('block', 'warn', 'none'));
