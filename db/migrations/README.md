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
8. billing
9. publicsite
