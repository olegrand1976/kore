-- Inaltérabilité du journal d'audit ETT (RG-ETT-01, critère PR-08.6) :
-- chaînage de hachage (tamper-evident) + interdiction stricte des UPDATE/DELETE.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE ett.audit_journal
    ADD COLUMN IF NOT EXISTS seq BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS prev_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS entry_hash TEXT NOT NULL DEFAULT '';

-- Backfill seq monotone par tenant sur les lignes existantes.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY tenant_id ORDER BY created_at, id) AS rn
    FROM ett.audit_journal
)
UPDATE ett.audit_journal j
SET seq = ranked.rn
FROM ranked
WHERE j.id = ranked.id;

-- Recalcule prev_hash / entry_hash pour l'historique (même algorithme que domain.ComputeHash).
DO $$
DECLARE
    cur_tenant UUID;
    v_prev TEXT;
    v_hash TEXT;
    r RECORD;
BEGIN
    cur_tenant := NULL;
    v_prev := '';
    FOR r IN
        SELECT id, tenant_id, seq, record_id, action, actor_id, created_at, payload
        FROM ett.audit_journal
        ORDER BY tenant_id, seq
    LOOP
        IF cur_tenant IS DISTINCT FROM r.tenant_id THEN
            cur_tenant := r.tenant_id;
            v_prev := '';
        END IF;
        v_hash := encode(
            digest(
                v_prev || '|' ||
                r.seq::TEXT || '|' ||
                r.tenant_id::TEXT || '|' ||
                r.record_id::TEXT || '|' ||
                r.action || '|' ||
                r.actor_id::TEXT || '|' ||
                to_char(r.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"') || '|' ||
                r.payload::TEXT,
                'sha256'
            ),
            'hex'
        );
        UPDATE ett.audit_journal
        SET prev_hash = v_prev, entry_hash = v_hash
        WHERE id = r.id;
        v_prev := v_hash;
    END LOOP;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ett_audit_tenant_seq
    ON ett.audit_journal(tenant_id, seq);

CREATE OR REPLACE FUNCTION ett.prevent_audit_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'ett.audit_journal is append-only (RECORD_IMMUTABLE)'
        USING ERRCODE = 'restrict_violation';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_ett_audit_no_mutation ON ett.audit_journal;
CREATE TRIGGER trg_ett_audit_no_mutation
    BEFORE UPDATE OR DELETE ON ett.audit_journal
    FOR EACH ROW EXECUTE FUNCTION ett.prevent_audit_mutation();
