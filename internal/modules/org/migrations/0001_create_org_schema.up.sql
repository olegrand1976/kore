CREATE SCHEMA IF NOT EXISTS org;

CREATE TABLE IF NOT EXISTS org.tenants (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org.societes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    raison_sociale TEXT NOT NULL,
    logo TEXT,
    devise TEXT NOT NULL DEFAULT 'EUR',
    langue_defaut TEXT NOT NULL DEFAULT 'fr',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org.sites (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    societe_id UUID NOT NULL REFERENCES org.societes(id),
    libelle TEXT NOT NULL,
    pays TEXT NOT NULL DEFAULT 'FR',
    strategie_budget TEXT NOT NULL DEFAULT 'standard',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org.services (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    site_id UUID NOT NULL REFERENCES org.sites(id),
    type TEXT NOT NULL DEFAULT 'interne',
    responsable_id UUID,
    commercial_id UUID,
    assistante_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org.applications (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    service_id UUID NOT NULL REFERENCES org.services(id),
    libelle TEXT NOT NULL,
    proprietaire TEXT,
    mode_facturation TEXT NOT NULL DEFAULT 'temps_passe',
    budget_defaut_id UUID,
    uo_activee BOOLEAN NOT NULL DEFAULT FALSE,
    chef_utilisateur_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org.equipes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    application_id UUID NOT NULL REFERENCES org.applications(id),
    libelle TEXT NOT NULL,
    responsable_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org.users (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    equipe_id UUID REFERENCES org.equipes(id),
    login TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    profil TEXT NOT NULL,
    langue TEXT NOT NULL DEFAULT 'fr',
    cra_requis BOOLEAN NOT NULL DEFAULT TRUE,
    type_compte TEXT NOT NULL DEFAULT 'Interne',
    salarie_ett BOOLEAN NOT NULL DEFAULT FALSE,
    date_activation DATE NOT NULL DEFAULT CURRENT_DATE,
    date_expiration DATE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, login)
);

CREATE TABLE IF NOT EXISTS org.clients (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES org.tenants(id),
    raison_sociale TEXT NOT NULL,
    tva TEXT,
    contacts JSONB NOT NULL DEFAULT '[]',
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS org.authx_permissions (
    profile TEXT NOT NULL,
    module TEXT NOT NULL,
    action TEXT NOT NULL,
    PRIMARY KEY (profile, module, action)
);

INSERT INTO org.authx_permissions (profile, module, action) VALUES
    ('Administrateur', 'org', 'L'), ('Administrateur', 'org', 'E'), ('Administrateur', 'org', 'V'),
    ('Administrateur', 'cra', 'L'), ('Administrateur', 'cra', 'E'), ('Administrateur', 'cra', 'V'),
    ('Administrateur', 'tma', 'L'), ('Administrateur', 'tma', 'E'), ('Administrateur', 'tma', 'V'),
    ('Collaborateur', 'cra', 'L'), ('Collaborateur', 'cra', 'E'),
    ('Chef d''équipe', 'cra', 'L'), ('Chef d''équipe', 'cra', 'V'),
    ('Utilisateur', 'tma', 'L'), ('Utilisateur', 'tma', 'E')
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_org_users_tenant ON org.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_org_clients_tenant ON org.clients(tenant_id);
