ALTER TABLE tma.demands
    ADD COLUMN IF NOT EXISTS taken_over_by_id UUID;
