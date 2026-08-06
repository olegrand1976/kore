# Déploiement GCP Premedica 2025

Kore s'intègre dans l'écosystème LL-IT sur le projet `premedica-prod-2025` (pattern Business Management / EquiMind).

Voir aussi `technical/foundation/09-gcp-infrastructure.md`.

## Ressources

| Ressource | Valeur |
|-----------|--------|
| Projet GCP | `premedica-prod-2025` |
| Cloud Run | `kore-api`, `kore-frontend` (europe-west9) |
| PostgreSQL | `premedica-db-staging` / base `kore` |
| Redis | VM `shared-redis`, DB **13**, préfixe `kore:` |
| Domaine | `kore.ll-it-sc.be` |
| Monitoring | Business Management `/admin/infra/monitor` |

## Première installation

```bash
# 1. Bootstrap (DB, users, secrets, SA, Artifact Registry)
make gcp-setup

# 2. WIF GitHub Actions (une fois)
make gcp-github-deploy

# 3. Déploiement complet
make gcp-deploy

# 4. Jobs + seed initial + smoke
make gcp-postdeploy-full

# 5. Domaine custom (LB + certificat)
make gcp-domain
# Puis DNS OVH : A kore → 34.54.99.89
```

## CI/CD

- **CI** : `.github/workflows/ci.yml` (tests sur chaque PR)
- **Deploy** : `.github/workflows/deploy-gcp-staging.yml` (push `staging` → Cloud Build → **smoke only**, sans seed-reset)
- **Main** : pas de déploiement GCP automatique pour l'instant
- **Wiki** : job `sync-wiki` — publie `documentation/`, `technical/` et `db/migrations/README.md` sur [le wiki du projet](https://github.com/olegrand1976/kore/wiki) via `scripts/sync-github-wiki.sh`

> **Production / société protégée** : dès qu'une société a `seed_protected = true` (via `make bootstrap-llit` / `make gcp-bootstrap-llit`), `seed-reset` et le job Cloud Run `kore-seed-reset` sont **refusés**. Ne jamais relancer `--seed-reset` sur un environnement LL-IT.
>
> Compte admin prod : `ADM_olivier` — créé en local (`make bootstrap-llit`) et à chaque deploy staging (`--bootstrap-llit`). Mot de passe : env `KORE_PROD_ADMIN_PASSWORD` ou secret GCP `kore-prod-admin-password` ; sinon généré une fois (logs du job).

Secret GitHub requis pour le wiki (le `GITHUB_TOKEN` ne peut pas pousser vers le dépôt `.wiki`) :

- `WIKI_SYNC_TOKEN` — PAT classic avec scope `repo` (ou fine-grained : Contents read/write sur ce dépôt)

Secrets GitHub (configurés via WIF, pas de clé JSON) :

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT` = `github-kore-deploy@premedica-prod-2025.iam.gserviceaccount.com`

## Commandes utiles

```bash
make gcp-deploy          # Rebuild + migrate + deploy API + frontend
make gcp-deploy-jobs     # Redéployer les Cloud Run Jobs uniquement
make gcp-postdeploy          # Smoke test (après deploy CI) — sans seed-reset
make gcp-postdeploy-staging  # Legacy : seed reset + smoke (interdit si société seed_protected)
make gcp-smoke           # Vérifier /health et /ready
make bootstrap-llit      # Local : société LL-IT protégée + ADM_olivier
```

> Ne pas utiliser `make gcp-postdeploy-staging` (seed-reset) sur un environnement avec organisation protégée.
## PDF CRA (Chromium)

L'image API (`deploy/Dockerfile.api`) embarque Chromium (`CHROME_PATH=/usr/bin/chromium`) pour la génération PDF CRA. En local sans Chrome, la génération PDF renvoie une erreur explicite (pas de fallback HTML).

## Infra partagée (repo `projets/infra`)

Kore est enregistré dans :

- `infra/database-backup-registry.yaml` (backups quotidiens PostgreSQL)
- `infra/shared-redis/redis-apps.conf` (DB 13)
- `infra/shared-postgres/setup-db-protection.sh` (grants `kore_app` / `kore_migrate`)

Après modification du registre infra :

```bash
cd ../infra
./shared-postgres/setup-backups.sh
./shared-redis/setup-gcp.sh
```

## Build images seules (legacy)

```bash
gcloud builds submit --config=deploy/cloudbuild.yaml \
  --substitutions=SHORT_SHA=$(git rev-parse --short HEAD)
```

Le déploiement complet utilise `infra/gcp/cloudbuild.yaml`.
