-- Allow multiple time lines with the same source on the same day (e.g. 2x prestation).
ALTER TABLE cra.time_lines
    DROP CONSTRAINT IF EXISTS time_lines_week_entry_id_source_type_source_id_day_key;
