DROP TRIGGER IF EXISTS trg_ett_audit_no_mutation ON ett.audit_journal;
DROP FUNCTION IF EXISTS ett.prevent_audit_mutation();
DROP INDEX IF EXISTS ett.idx_ett_audit_tenant_seq;
ALTER TABLE ett.audit_journal
    DROP COLUMN IF EXISTS entry_hash,
    DROP COLUMN IF EXISTS prev_hash,
    DROP COLUMN IF EXISTS seq;
