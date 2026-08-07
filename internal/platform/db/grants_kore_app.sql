-- Privileges for kore_app (idempotent). Applied automatically after migrations
-- by kore-api migrate. No-op when role kore_app is absent (local docker user "kore").
-- DEFAULT PRIVILEGES target kore_migrate when present, else current_user.

DO $grants$
DECLARE
  sch text;
  schemas text[] := ARRAY[
    'public', 'platform', 'org', 'cra', 'conges', 'budget', 'tma', 'billing',
    'workflow', 'notifications', 'publicsite', 'ai', 'ssii', 'support',
    'maintenance', 'invoicing', 'ett', 'reporting', 'admin', 'integrations'
  ];
  creator text;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'kore_app') THEN
    RAISE NOTICE 'kore_app role absent — skipping app grants';
    RETURN;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'kore_migrate') THEN
    creator := 'kore_migrate';
  ELSE
    creator := current_user;
  END IF;

  FOREACH sch IN ARRAY schemas LOOP
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = sch) THEN
      CONTINUE;
    END IF;
    EXECUTE format('GRANT USAGE ON SCHEMA %I TO kore_app', sch);
    EXECUTE format(
      'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %I TO kore_app',
      sch
    );
    EXECUTE format(
      'GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %I TO kore_app',
      sch
    );
    EXECUTE format(
      'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA %I GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO kore_app',
      creator, sch
    );
    EXECUTE format(
      'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA %I GRANT USAGE, SELECT ON SEQUENCES TO kore_app',
      creator, sch
    );
  END LOOP;

  -- ETT audit journal is append-only (trigger enforces; revoke is defense in depth).
  IF to_regclass('ett.audit_journal') IS NOT NULL THEN
    EXECUTE 'REVOKE UPDATE, DELETE ON TABLE ett.audit_journal FROM kore_app';
  END IF;
END
$grants$;
