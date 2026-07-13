#!/usr/bin/env bash
# Bootstrap GCP Kore sur premedica-prod-2025 (DB, secrets, SA, Artifact Registry).
# Usage: ./infra/gcp/setup-gcp.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== Kore setup GCP — ${GCP_PROJECT_ID} ==="

ensure_secret() {
  local name="$1"
  if ! gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "→ CREATE secret ${name}"
    gcloud secrets create "$name" --replication-policy=automatic --project="$GCP_PROJECT_ID" --quiet
  fi
}

add_secret_version() {
  local name="$1"
  local value="$2"
  ensure_secret "$name"
  echo -n "$value" | gcloud secrets versions add "$name" --data-file=- --project="$GCP_PROJECT_ID" --quiet
}

SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
CB_SA="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')@cloudbuild.gserviceaccount.com"

echo "→ IAM Cloud Build"
for role in roles/run.admin roles/artifactregistry.writer roles/iam.serviceAccountUser roles/secretmanager.secretAccessor roles/cloudsql.admin; do
  gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
    --member="serviceAccount:${CB_SA}" --role="$role" --quiet >/dev/null 2>&1 || true
done

if ! gcloud iam service-accounts describe "$SA_EMAIL" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ CREATE service account ${SERVICE_ACCOUNT}"
  gcloud iam service-accounts create "$SERVICE_ACCOUNT" \
    --display-name="Kore Cloud Run" --project="$GCP_PROJECT_ID" --quiet
fi

gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --member="serviceAccount:${CB_SA}" --role="roles/iam.serviceAccountUser" --quiet >/dev/null 2>&1 || true

for role in roles/cloudsql.client roles/secretmanager.secretAccessor roles/run.invoker; do
  gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" --role="$role" --quiet >/dev/null 2>&1 || true
done

if ! gcloud artifacts repositories describe "$AR_REPO" --location="$GCP_AR_REGION" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ CREATE Artifact Registry ${AR_REPO}"
  gcloud artifacts repositories create "$AR_REPO" \
    --repository-format=docker --location="$GCP_AR_REGION" --project="$GCP_PROJECT_ID" --quiet
else
  echo "  Artifact Registry ${AR_REPO} existe"
fi

echo "→ PostgreSQL (${SQL_INSTANCE})"
if ! gcloud sql databases describe "$DB_NAME" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud sql databases create "$DB_NAME" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --quiet
else
  echo "  Base ${DB_NAME} existe"
fi

if ! gcloud sql users list --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --format='value(name)' | grep -qx "$DB_USER"; then
  DB_PASS="${KORE_DB_PASSWORD:-$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)}"
  echo "→ CREATE USER ${DB_USER}"
  gcloud sql users create "$DB_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$DB_PASS" --quiet
  add_secret_version "kore-db-password" "$DB_PASS"
else
  DB_PASS="$(gcloud secrets versions access latest --secret=kore-db-password --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$DB_PASS" ]]; then
    echo "ERREUR: utilisateur ${DB_USER} existe mais secret kore-db-password vide" >&2
    exit 1
  fi
  echo "  Utilisateur ${DB_USER} existe"
fi

ENC_PASS="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$DB_PASS")"
DATABASE_URL="postgres://${DB_USER}:${ENC_PASS}@/${DB_NAME}?host=/cloudsql/${CLOUDSQL_INSTANCE}"
add_secret_version "kore-database-url" "$DATABASE_URL"
echo "→ kore-database-url mis à jour (runtime ${DB_USER})"

if ! gcloud sql users list --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --format='value(name)' | grep -qx "$MIGRATE_USER"; then
  MIGRATE_PASS="${KORE_MIGRATE_DB_PASSWORD:-$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)}"
  echo "→ CREATE USER ${MIGRATE_USER}"
  gcloud sql users create "$MIGRATE_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$MIGRATE_PASS" --quiet
  add_secret_version "kore-migrate-db-password" "$MIGRATE_PASS"
else
  MIGRATE_PASS="$(gcloud secrets versions access latest --secret=kore-migrate-db-password --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$MIGRATE_PASS" ]]; then
    MIGRATE_PASS="$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)"
    echo "→ RESET secret migrate (user ${MIGRATE_USER} existait sans secret)"
    gcloud sql users set-password "$MIGRATE_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$MIGRATE_PASS" --quiet
    add_secret_version "kore-migrate-db-password" "$MIGRATE_PASS"
  else
    echo "  Utilisateur ${MIGRATE_USER} existe"
  fi
fi

ENC_MIGRATE_PASS="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$MIGRATE_PASS")"
MIGRATE_DATABASE_URL="postgres://${MIGRATE_USER}:${ENC_MIGRATE_PASS}@/${DB_NAME}?host=/cloudsql/${CLOUDSQL_INSTANCE}"
add_secret_version "kore-migrate-database-url" "$MIGRATE_DATABASE_URL"
echo "→ kore-migrate-database-url mis à jour (${MIGRATE_USER})"

if ! gcloud secrets versions list kore-jwt-signing-key --project="$GCP_PROJECT_ID" --limit=1 --format='value(name)' 2>/dev/null | grep -q .; then
  JWT_KEY="$(openssl rand -base64 48 | tr -d '/+=' | head -c 64)"
  add_secret_version "kore-jwt-signing-key" "$JWT_KEY"
  echo "→ kore-jwt-signing-key créé"
else
  echo "  kore-jwt-signing-key existe"
fi

for stripe_secret in kore-stripe-secret-key kore-stripe-webhook-secret kore-stripe-publishable-key; do
  ensure_secret "$stripe_secret"
  if ! gcloud secrets versions list "$stripe_secret" --project="$GCP_PROJECT_ID" --limit=1 --format='value(name)' 2>/dev/null | grep -q .; then
    case "$stripe_secret" in
      kore-stripe-secret-key) add_secret_version "$stripe_secret" "sk_test_placeholder" ;;
      kore-stripe-webhook-secret) add_secret_version "$stripe_secret" "whsec_placeholder" ;;
      kore-stripe-publishable-key) add_secret_version "$stripe_secret" "pk_test_placeholder" ;;
    esac
    echo "→ ${stripe_secret} placeholder — remplacez par les vraies clés Stripe"
  else
    echo "  ${stripe_secret} existe"
  fi
done

COMPUTE_SA="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')-compute@developer.gserviceaccount.com"
for secret in kore-database-url kore-migrate-database-url kore-jwt-signing-key kore-redis-url \
  kore-stripe-secret-key kore-stripe-webhook-secret kore-stripe-publishable-key; do
  if gcloud secrets describe "$secret" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    gcloud secrets add-iam-policy-binding "$secret" \
      --project="$GCP_PROJECT_ID" \
      --member="serviceAccount:${SA_EMAIL}" \
      --role=roles/secretmanager.secretAccessor \
      --quiet >/dev/null 2>&1 || true
    gcloud secrets add-iam-policy-binding "$secret" \
      --project="$GCP_PROJECT_ID" \
      --member="serviceAccount:${COMPUTE_SA}" \
      --role=roles/secretmanager.secretAccessor \
      --quiet >/dev/null 2>&1 || true
  fi
done

if [[ -d "${INFRA_ROOT}/shared-postgres" ]]; then
  echo "→ Privilèges PostgreSQL kore_app (idempotent)"
  bash "${INFRA_ROOT}/shared-postgres/setup-db-protection.sh" 2>&1 | tail -5 || true
fi

if [[ -d "${INFRA_ROOT}/shared-redis" ]]; then
  echo "→ Sync secrets Redis (infra partagée)"
  bash "${INFRA_ROOT}/shared-redis/setup-gcp.sh" 2>&1 | tail -8 || true
fi

cat <<EOF

Prêt pour :
  ./infra/gcp/setup-github-deploy.sh
  gcloud builds submit --config=infra/gcp/cloudbuild.yaml --project=${GCP_PROJECT_ID}
  ./infra/gcp/postdeploy.sh
  ./infra/gcp/setup-custom-domain.sh

Secrets Stripe placeholder à remplacer avant prod réelle.

EOF
