DROP TABLE IF EXISTS project.kanban_configs;
ALTER TABLE IF EXISTS project.epics DROP CONSTRAINT IF EXISTS project_epics_target_sprint_fk;
DROP TABLE IF EXISTS project.sprints;
DROP TABLE IF EXISTS project.epics;
DROP SCHEMA IF EXISTS project;
