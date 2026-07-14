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
- **Deploy** : `.github/workflows/deploy-gcp.yml` (push `staging` → Cloud Build → seed reset → smoke sur `kore.ll-it-sc.be`)
- **Main** : pas de déploiement GCP automatique pour l'instant
- **Wiki** : job `sync-wiki` — publie `documentation/`, `technical/` et `db/migrations/README.md` sur [le wiki du projet](https://github.com/olegrand1976/kore/wiki) via `scripts/sync-github-wiki.sh`

Secret GitHub requis pour le wiki (le `GITHUB_TOKEN` ne peut pas pousser vers le dépôt `.wiki`) :

- `WIKI_SYNC_TOKEN` — PAT classic avec scope `repo` (ou fine-grained : Contents read/write sur ce dépôt)

Secrets GitHub (configurés via WIF, pas de clé JSON) :

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_SERVICE_ACCOUNT` = `github-kore-deploy@premedica-prod-2025.iam.gserviceaccount.com`

## Commandes utiles

```bash
make gcp-deploy          # Rebuild + migrate + deploy API + frontend
make gcp-deploy-jobs     # Redéployer les Cloud Run Jobs uniquement
make gcp-postdeploy          # Smoke test (après deploy CI)
make gcp-postdeploy-staging  # Seed reset + smoke (staging)
make gcp-smoke           # Vérifier /health et /ready
```

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
