# Déploiement GCP (Phase F)

Scaffolding Cloud Build pour Kore. Voir `technical/foundation/09-gcp-infrastructure.md` pour l'architecture cible.

## Prérequis

- Projet GCP avec Artifact Registry (`kore` repository)
- Cloud SQL PostgreSQL + Memorystore Redis provisionnés
- Secret Manager : `JWT_SIGNING_KEY`, `DATABASE_URL`, `STRIPE_*`

## Build images

```bash
gcloud builds submit --config deploy/cloudbuild.yaml \
  --substitutions=SHORT_SHA=$(git rev-parse --short HEAD)
```

## Déploiement Cloud Run (manuel MVP)

1. Migrer la base via Cloud Run Job (`go run ./cmd/kore-api migrate`)
2. Déployer l'API : image `kore-api`, port 8080, VPC connector pour Cloud SQL/Redis
3. Déployer le frontend : image `kore-frontend`, port 3000, `NUXT_PUBLIC_API_BASE` vers l'URL API

## CI locale

GitHub Actions (`.github/workflows/ci.yml`) valide build + tests avant tout déploiement GCP.
