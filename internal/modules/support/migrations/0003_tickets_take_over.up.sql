ALTER TABLE support.tickets
    ADD COLUMN IF NOT EXISTS taken_over_by_id UUID;