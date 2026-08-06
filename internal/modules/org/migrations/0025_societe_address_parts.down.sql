ALTER TABLE org.societes
    DROP COLUMN IF EXISTS ville,
    DROP COLUMN IF EXISTS code_postal,
    DROP COLUMN IF EXISTS adresse_boite,
    DROP COLUMN IF EXISTS adresse_numero;
