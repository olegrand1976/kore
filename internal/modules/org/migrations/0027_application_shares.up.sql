-- Applications partagées : rattachements N–N sites / services / équipes (sans service propriétaire unique).

CREATE TABLE IF NOT EXISTS org.application_sites (
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    application_id UUID NOT NULL REFERENCES org.applications(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES org.sites(id) ON DELETE CASCADE,
    PRIMARY KEY (application_id, site_id)
);
CREATE INDEX IF NOT EXISTS idx_org_application_sites_site
    ON org.application_sites (tenant_id, site_id);

CREATE TABLE IF NOT EXISTS org.application_services (
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    application_id UUID NOT NULL REFERENCES org.applications(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES org.services(id) ON DELETE CASCADE,
    PRIMARY KEY (application_id, service_id)
);
CREATE INDEX IF NOT EXISTS idx_org_application_services_service
    ON org.application_services (tenant_id, service_id);

CREATE TABLE IF NOT EXISTS org.application_equipes (
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    application_id UUID NOT NULL REFERENCES org.applications(id) ON DELETE CASCADE,
    equipe_id UUID NOT NULL REFERENCES org.equipes(id) ON DELETE CASCADE,
    PRIMARY KEY (application_id, equipe_id)
);
CREATE INDEX IF NOT EXISTS idx_org_application_equipes_equipe
    ON org.application_equipes (tenant_id, equipe_id);

-- Backfill depuis l'ancien propriétaire unique service_id.
INSERT INTO org.application_services (tenant_id, application_id, service_id)
SELECT tenant_id, id, service_id
FROM org.applications
WHERE service_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Backfill home équipes → partages.
INSERT INTO org.application_equipes (tenant_id, application_id, equipe_id)
SELECT tenant_id, application_id, id
FROM org.equipes
WHERE application_id IS NOT NULL
ON CONFLICT DO NOTHING;

ALTER TABLE org.applications DROP COLUMN IF EXISTS service_id;
