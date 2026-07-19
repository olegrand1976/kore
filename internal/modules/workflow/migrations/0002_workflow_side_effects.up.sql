ALTER TABLE workflow.states
    ADD COLUMN IF NOT EXISTS on_enter_effects JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE workflow.transitions
    ADD COLUMN IF NOT EXISTS on_fire_effects JSONB NOT NULL DEFAULT '[]'::jsonb;
