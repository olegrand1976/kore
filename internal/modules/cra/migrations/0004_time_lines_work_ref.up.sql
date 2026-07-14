ALTER TABLE cra.time_lines
    ADD COLUMN IF NOT EXISTS work_ref_type TEXT,
    ADD COLUMN IF NOT EXISTS work_ref_id TEXT;
