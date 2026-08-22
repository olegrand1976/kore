#!/usr/bin/env bash
# Complète la configuration ops Taiga (secrets URL/slug + instructions webhook).
# Usage:
#   TAIGA_BASE_URL=https://taiga.example.com TAIGA_PROJECT_SLUG=mon-projet ./scripts/taiga-ops-complete.sh
# Optional: DEMAND_UUID pour le smoke (sinon première demande TMA via API).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

: "${TAIGA_BASE_URL:?TAIGA_BASE_URL requis}"
: "${TAIGA_PROJECT_SLUG:?TAIGA_PROJECT_SLUG requis}"

export TAIGA_BASE_URL TAIGA_PROJECT_SLUG
export TAIGA_DEFAULT_TENANT_ID="${TAIGA_DEFAULT_TENANT_ID:-00000000-0000-4000-8000-0000000000a1}"

echo "=== 1/3 Secrets GCP ==="
"${ROOT}/scripts/taiga-setup-gcp.sh"

echo ""
echo "=== 2/3 Redeploy API ==="
echo "→ Les secrets URL/slug sont lus au prochain deploy Cloud Run."
echo "  Push sur staging ou: make gcp-deploy"

echo ""
echo "=== 3/3 Configuration Taiga (manuel) ==="
echo "1. Admin Taiga → Settings → Webhooks"
echo "   URL    : https://kore.ll-it-sc.be/api/v1/integrations/taiga/webhook"
echo "   Secret : récupérer via gcloud secrets versions access latest --secret=kore-taiga-webhook-secret"
echo "2. Sur une user story : external_reference = [\"kore\", \"<uuid-demande-tma>\"]"
echo ""
echo "=== Smoke (après redeploy) ==="
if [[ -n "${DEMAND_UUID:-}" ]]; then
  "${ROOT}/scripts/taiga-webhook-smoke.sh" "https://kore.ll-it-sc.be" "${DEMAND_UUID}"
else
  echo "→ Après redeploy : ./scripts/taiga-webhook-smoke.sh https://kore.ll-it-sc.be <uuid-demande>"
fi
