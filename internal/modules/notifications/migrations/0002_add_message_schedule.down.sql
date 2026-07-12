DROP INDEX IF EXISTS notifications.idx_notifications_messages_due;

ALTER TABLE notifications.messages
    DROP COLUMN IF EXISTS scheduled_for;
