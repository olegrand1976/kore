ALTER TABLE cra.time_lines
    ADD CONSTRAINT time_lines_week_entry_id_source_type_source_id_day_key
    UNIQUE (week_entry_id, source_type, source_id, day);
