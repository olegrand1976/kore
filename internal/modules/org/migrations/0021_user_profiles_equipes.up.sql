-- Multi-profile / multi-team memberships (N–N).
-- Primary columns org.users.profil and org.users.equipe_id remain denormalized
-- (first / highest-privilege profile, first team) for JWT and legacy joins.

CREATE TABLE IF NOT EXISTS org.user_profiles (
    user_id UUID NOT NULL REFERENCES org.users(id) ON DELETE CASCADE,
    profil  TEXT NOT NULL,
    PRIMARY KEY (user_id, profil)
);

CREATE INDEX IF NOT EXISTS idx_org_user_profiles_profil
    ON org.user_profiles (profil);

CREATE TABLE IF NOT EXISTS org.user_equipes (
    user_id   UUID NOT NULL REFERENCES org.users(id) ON DELETE CASCADE,
    equipe_id UUID NOT NULL REFERENCES org.equipes(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, equipe_id)
);

CREATE INDEX IF NOT EXISTS idx_org_user_equipes_equipe
    ON org.user_equipes (equipe_id);

INSERT INTO org.user_profiles (user_id, profil)
SELECT id, profil FROM org.users
WHERE deleted_at IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO org.user_equipes (user_id, equipe_id)
SELECT id, equipe_id FROM org.users
WHERE equipe_id IS NOT NULL AND deleted_at IS NULL
ON CONFLICT DO NOTHING;
