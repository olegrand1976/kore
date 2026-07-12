ALTER TABLE notifications.messages
    ADD COLUMN IF NOT EXISTS scheduled_for TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_notifications_messages_due
    ON notifications.messages(status, scheduled_for);
