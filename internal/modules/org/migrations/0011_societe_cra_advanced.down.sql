ALTER TABLE org.societes
  DROP COLUMN IF EXISTS week_submit_policy,
  DROP COLUMN IF EXISTS cra_mail_recipients,
  DROP COLUMN IF EXISTS cra_mail_auto,
  DROP COLUMN IF EXISTS day_capacity_minutes;
