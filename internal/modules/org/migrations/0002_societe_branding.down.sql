ALTER TABLE org.societes
    DROP COLUMN IF EXISTS url_tenant,
    DROP COLUMN IF EXISTS siret,
    DROP COLUMN IF EXISTS adresse;
