#!/usr/bin/env bash
# Smoke test webhook Taiga (local ou staging).
# Usage: ./scripts/taiga-webhook-smoke.sh [BASE_URL] [DEMAND_UUID]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=../infra/gcp/lib/gcp-env.sh
source "${ROOT}/infra/gcp/lib/gcp-env.sh"

BASE_URL="${1:-https://${CUSTOM_DOMAIN}}"
DEMAND_ID="${2:-}"

if [[ -z "$DEMAND_ID" ]]; then
  DEMAND_ID="$(python3 -c 'import uuid; print(uuid.uuid4())')"
  echo "→ demand_id généré (fictif) : ${DEMAND_ID}"
fi

SECRET="${TAIGA_WEBHOOK_SECRET:-}"
if [[ -z "$SECRET" ]]; then
  if gcloud secrets versions access latest --secret=kore-taiga-webhook-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    SECRET="$(gcloud secrets versions access latest --secret=kore-taiga-webhook-secret --project="$GCP_PROJECT_ID")"
  else
    echo "TAIGA_WEBHOOK_SECRET ou secret GCP kore-taiga-webhook-secret requis" >&2
    exit 1
  fi
fi

TENANT="${TAIGA_DEFAULT_TENANT_ID:-00000000-0000-4000-8000-0000000000a1}"
WEBHOOK_URL="${BASE_URL%/}/api/v1/integrations/taiga/webhook"

payload=$(cat <<EOF
{
  "action": "create",
  "type": "userstory",
  "data": {
    "id": 9001,
    "ref": 42,
    "project": { "slug": "kore-demo" },
    "external_reference": ["kore", "${DEMAND_ID}"],
    "permalink": "https://taiga.example/project/kore-demo/us/42"
  }
}
EOF
)

echo "→ POST ${WEBHOOK_URL}"
http_code=$(curl -sS -o /tmp/taiga-smoke-body.txt -w '%{http_code}' \
  -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Taiga-Webhook-Secret: ${SECRET}" \
  -H "X-Kore-Tenant-ID: ${TENANT}" \
  -d "$payload")

echo "← HTTP ${http_code}"
cat /tmp/taiga-smoke-body.txt
echo ""

if [[ "$http_code" != "200" ]]; then
  exit 1
fi

echo "→ GET lien (nécessite session — test manuel sur fiche TMA /tma/${DEMAND_ID})"
