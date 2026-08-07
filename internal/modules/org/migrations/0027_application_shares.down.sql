-- Restore unique service_id owner (first linked service if any).
ALTER TABLE org.applications ADD COLUMN IF NOT EXISTS service_id UUID REFERENCES org.services(id);

UPDATE org.applications a
SET service_id = s.service_id
FROM (
    SELECT DISTINCT ON (application_id) application_id, service_id
    FROM org.application_services
    ORDER BY application_id, service_id
) s
WHERE a.id = s.application_id AND a.service_id IS NULL;

-- Remaining apps without a service share cannot satisfy NOT NULL; leave nullable if empty.
-- Best-effort: apps that still lack service_id keep NULL (down is best-effort).

DROP TABLE IF EXISTS org.application_equipes;
DROP TABLE IF EXISTS org.application_services;
DROP TABLE IF EXISTS org.application_sites;
