# Migrations

Les migrations SQL vivent dans `internal/modules/<module>/migrations/` et sont appliquées par le runner Go maison (`platform/db`) via `kore-api migrate`.

Ordre d'application (cf. `internal/app/migrations.go`) :

1. org
2. workflow
3. cra
4. notifications
5. conges
6. budget
7. tma
8. ai
9. billing
10. publicsite

**Schéma DB documenté** : [`documentation/SCHEMA_DB.md`](../../documentation/SCHEMA_DB.md)
