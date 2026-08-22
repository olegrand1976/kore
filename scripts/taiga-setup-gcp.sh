#!/usr/bin/env bash
# Bootstrap secrets GCP pour l'intégration Taiga (webhook + URLs).
# Usage: ./scripts/taiga-setup-gcp.sh [--rotate-webhook]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=../infra/gcp/lib/gcp-env.sh
source "${ROOT}/infra/gcp/lib/gcp-env.sh"

ROTATE_WEBHOOK=false
for arg in "$@"; do
  case "$arg" in
    --rotate-webhook) ROTATE_WEBHOOK=true ;;
    -h | --help)
      echo "Usage: $0 [--rotate-webhook]"
      exit 0
      ;;
    *)
      echo "Option inconnue: $arg" >&2
      exit 1
      ;;
  esac
done

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

ensure_secret() {
  local name="$1"
  if ! gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "→ CREATE secret ${name}"
    gcloud secrets create "$name" --replication-policy=automatic --project="$GCP_PROJECT_ID" --quiet
  fi
}

add_version_if_missing() {
  local name="$1"
  local value="$2"
  ensure_secret "$name"
  if gcloud secrets versions list "$name" --project="$GCP_PROJECT_ID" --limit=1 --format='value(name)' 2>/dev/null | grep -q .; then
    echo "  ${name} : version existante (inchangée)"
    return 0
  fi
  echo -n "$value" | gcloud secrets versions add "$name" --data-file=- --project="$GCP_PROJECT_ID" --quiet
  echo "  ${name} : version initiale ajoutée"
}

add_version_force() {
  local name="$1"
  local value="$2"
  ensure_secret "$name"
  echo -n "$value" | gcloud secrets versions add "$name" --data-file=- --project="$GCP_PROJECT_ID" --quiet
  echo "  ${name} : nouvelle version ajoutée"
}

has_version() {
  gcloud secrets versions list "$1" --project="$GCP_PROJECT_ID" --limit=1 --format='value(name)' 2>/dev/null | grep -q .
}

echo "=== Taiga secrets — ${GCP_PROJECT_ID} ==="

WEBHOOK_SECRET="${TAIGA_WEBHOOK_SECRET:-}"
if [[ -z "$WEBHOOK_SECRET" ]]; then
  WEBHOOK_SECRET="$(openssl rand -hex 32)"
fi

if $ROTATE_WEBHOOK || ! has_version kore-taiga-webhook-secret; then
  if $ROTATE_WEBHOOK; then
    add_version_force kore-taiga-webhook-secret "$WEBHOOK_SECRET"
  else
    add_version_if_missing kore-taiga-webhook-secret "$WEBHOOK_SECRET"
  fi
else
  WEBHOOK_SECRET="$(gcloud secrets versions access latest --secret=kore-taiga-webhook-secret --project="$GCP_PROJECT_ID")"
  echo "  kore-taiga-webhook-secret : version existante conservée"
fi

DEFAULT_TENANT="${TAIGA_DEFAULT_TENANT_ID:-00000000-0000-4000-8000-0000000000a1}"
add_version_if_missing kore-taiga-default-tenant-id "$DEFAULT_TENANT"

for cfg in kore-taiga-base-url kore-taiga-project-slug; do
  ensure_secret "$cfg"
  if has_version "$cfg"; then
    echo "  ${cfg} : version existante"
  else
    echo "  ${cfg} : secret créé — ajoutez une version (voir ci-dessous)"
  fi
done

if [[ -n "${TAIGA_BASE_URL:-}" ]]; then
  add_version_force kore-taiga-base-url "$TAIGA_BASE_URL"
fi
if [[ -n "${TAIGA_PROJECT_SLUG:-}" ]]; then
  add_version_force kore-taiga-project-slug "$TAIGA_PROJECT_SLUG"
fi

SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
COMPUTE_SA="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')-compute@developer.gserviceaccount.com"
for secret in kore-taiga-webhook-secret kore-taiga-base-url kore-taiga-project-slug kore-taiga-default-tenant-id; do
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
done

echo ""
echo "=== Configuration webhook Taiga ==="
echo "URL     : https://${CUSTOM_DOMAIN}/api/v1/integrations/taiga/webhook"
echo "Header  : X-Taiga-Webhook-Secret: ${WEBHOOK_SECRET}"
echo "Option  : X-Kore-Tenant-ID: ${DEFAULT_TENANT} (si mono-tenant)"
echo ""
echo "URLs Taiga (si absentes) :"
echo "  echo -n 'https://taiga.example.com' | gcloud secrets versions add kore-taiga-base-url --data-file=- --project=${GCP_PROJECT_ID}"
echo "  echo -n 'mon-projet' | gcloud secrets versions add kore-taiga-project-slug --data-file=- --project=${GCP_PROJECT_ID}"
echo ""
echo "Redéployez l'API (push staging ou make gcp-deploy) pour monter les secrets."
