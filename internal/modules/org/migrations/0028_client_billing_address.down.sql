ALTER TABLE org.clients
    DROP COLUMN IF EXISTS pays,
    DROP COLUMN IF EXISTS adresse,
    DROP COLUMN IF EXISTS adresse_numero,
    DROP COLUMN IF EXISTS adresse_boite,
    DROP COLUMN IF EXISTS code_postal,
    DROP COLUMN IF EXISTS ville,
    DROP COLUMN IF EXISTS siret;
