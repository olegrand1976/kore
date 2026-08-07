ALTER TABLE workflow.instances
    ADD COLUMN IF NOT EXISTS requester_id UUID;
