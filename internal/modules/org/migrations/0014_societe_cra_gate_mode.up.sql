ALTER TABLE org.societes
  ADD COLUMN IF NOT EXISTS cra_gate_mode TEXT NOT NULL DEFAULT 'warn';
