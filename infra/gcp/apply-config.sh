#!/usr/bin/env bash
# Publie la clé API Gemini (Secret Manager) depuis .env.gemini
# Usage : ./infra/gcp/apply-config.sh [--local-only]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

LOCAL_ONLY=false
if [[ "${1:-}" == "--local-only" ]]; then
  LOCAL_ONLY=true
fi

load_gemini_env() {
  local file="$1"
  if [[ -f "$file" ]]; then
    # shellcheck disable=SC1090
    set -a && source "$file" && set +a
  fi
}

load_gemini_env "${REPO_ROOT}/.env.gemini"
# Rétrocompat : ancien fichier .env.gcp
if [[ -z "${GEMINI_API_KEY:-}" ]]; then
  load_gemini_env "${REPO_ROOT}/.env.gcp"
  if [[ -n "${GCP_API_KEY:-}" && -z "${GEMINI_API_KEY:-}" ]]; then
    GEMINI_API_KEY="$GCP_API_KEY"
  fi
fi

if [[ -z "${GEMINI_API_KEY:-}" ]]; then
  echo "ERREUR: GEMINI_API_KEY manquant — renseignez ${REPO_ROOT}/.env.gemini" >&2
  exit 1
fi

echo "=== Kore Gemini apply-config — ${GCP_PROJECT_ID} (${GCP_PROJECT_NUMBER}) ==="

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

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
  echo "→ ${name} mis à jour"
}

if [[ "$LOCAL_ONLY" == "true" ]]; then
  echo "→ Mode local uniquement — gcloud project=${GCP_PROJECT_ID}"
  echo "  Clé Gemini présente (${#GEMINI_API_KEY} caractères)"
  exit 0
fi

add_secret_version "$GEMINI_API_KEY_SECRET" "$GEMINI_API_KEY"

SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
COMPUTE_SA="${GCP_PROJECT_NUMBER}-compute@developer.gserviceaccount.com"
for member in "serviceAccount:${SA_EMAIL}" "serviceAccount:${COMPUTE_SA}"; do
  gcloud secrets add-iam-policy-binding "$GEMINI_API_KEY_SECRET" \
    --project="$GCP_PROJECT_ID" \
    --member="$member" \
    --role=roles/secretmanager.secretAccessor \
    --quiet >/dev/null 2>&1 || true
done

cat <<EOF

Configuration Gemini appliquée :
  Projet     : ${GCP_PROJECT_ID} (${GCP_PROJECT_NUMBER})
  Secret     : ${GEMINI_API_KEY_SECRET}
  Provider   : AI_LLM_PROVIDER=gemini
  Domaine    : ${CUSTOM_DOMAIN}

Prochaines étapes :
  make gcp-deploy          # déployer API + frontend
  make gcp-postdeploy-full # migrate + seed + smoke

EOF
