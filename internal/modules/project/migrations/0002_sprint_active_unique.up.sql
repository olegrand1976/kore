CREATE UNIQUE INDEX IF NOT EXISTS idx_project_sprints_one_active
    ON project.sprints (tenant_id, application_id)
    WHERE status = 'active';
