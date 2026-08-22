#!/usr/bin/env bash
# Smoke test webhook Taiga (local ou staging).
# Usage: ./scripts/taiga-webhook-smoke.sh [BASE_URL] [DEMAND_UUID]
# Optional: SMOKE_LOGIN / SMOKE_PASSWORD for GET link verification (default seed admin).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=../infra/gcp/lib/gcp-env.sh
source "${ROOT}/infra/gcp/lib/gcp-env.sh"

BASE_URL="${1:-https://${CUSTOM_DOMAIN}}"
DEMAND_ID="${2:-}"
BODY_FILE="$(mktemp)"
trap 'rm -f "$BODY_FILE"' EXIT

API_BASE="${BASE_URL%/}/api/v1"

resolve_demand_id() {
  if [[ -n "$DEMAND_ID" ]]; then
    return 0
  fi
  local login="${SMOKE_LOGIN:-ADM_admin}"
  local password="${SMOKE_PASSWORD:-Admin123!}"
  local login_resp token demands

  echo "→ Résolution demand_id via API (première demande TMA)…"
  if ! login_resp=$(curl -sS -f -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"login\":\"${login}\",\"password\":\"${password}\"}" 2>/dev/null); then
    echo "ERREUR: login impossible — passez DEMAND_UUID en 2e argument" >&2
    exit 1
  fi
  token=$(echo "$login_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('accessToken') or d.get('AccessToken') or '')")
  if [[ -z "$token" ]]; then
    echo "ERREUR: token absent" >&2
    exit 1
  fi
  if ! demands=$(curl -sS -f -H "Authorization: Bearer ${token}" "${API_BASE}/demands?page_size=1" 2>/dev/null); then
    echo "ERREUR: GET /demands impossible" >&2
    exit 1
  fi
  DEMAND_ID=$(echo "$demands" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or []; print((d[0].get('id') or d[0].get('ID') or '') if d else '')")
  if [[ -z "$DEMAND_ID" ]]; then
    echo "ERREUR: aucune demande TMA — créez-en une ou passez DEMAND_UUID" >&2
    exit 1
  fi
  echo "→ demand_id : ${DEMAND_ID}"
}

resolve_demand_id

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
WEBHOOK_URL="${API_BASE}/integrations/taiga/webhook"
SMOKE_STORY_ID="${TAIGA_SMOKE_STORY_ID:-$((9000 + RANDOM % 9000))}"

payload=$(cat <<EOF
{
  "action": "create",
  "type": "userstory",
  "data": {
    "id": ${SMOKE_STORY_ID},
    "ref": 42,
    "project": { "slug": "kore-demo" },
    "external_reference": ["kore", "${DEMAND_ID}"],
    "permalink": "https://taiga.example/project/kore-demo/us/42"
  }
}
EOF
)

echo "→ POST ${WEBHOOK_URL}"
http_code=$(curl -sS -o "$BODY_FILE" -w '%{http_code}' \
  -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Taiga-Webhook-Secret: ${SECRET}" \
  -H "X-Kore-Tenant-ID: ${TENANT}" \
  -d "$payload")

echo "← HTTP ${http_code}"
cat "$BODY_FILE"
echo ""

if [[ "$http_code" != "200" ]]; then
  exit 1
fi

verify_link_get() {
  local login="${SMOKE_LOGIN:-ADM_admin}"
  local password="${SMOKE_PASSWORD:-Admin123!}"
  local login_resp token get_code

  echo "→ POST ${API_BASE}/auth/login (${login})"
  if ! login_resp=$(curl -sS -f -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"login\":\"${login}\",\"password\":\"${password}\"}" 2>/dev/null); then
    echo "→ GET lien ignoré (login impossible — définir SMOKE_LOGIN/SMOKE_PASSWORD)" >&2
    return 0
  fi

  token=$(echo "$login_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('accessToken') or d.get('AccessToken') or '')")
  if [[ -z "$token" ]]; then
    echo "→ GET lien ignoré (token absent dans la réponse login)" >&2
    return 0
  fi

  echo "→ GET ${API_BASE}/integrations/taiga/links/by-demand/${DEMAND_ID}"
  get_code=$(curl -sS -o "$BODY_FILE" -w '%{http_code}' \
    -H "Authorization: Bearer ${token}" \
    "${API_BASE}/integrations/taiga/links/by-demand/${DEMAND_ID}")

  echo "← HTTP ${get_code}"
  cat "$BODY_FILE"
  echo ""

  if [[ "$get_code" != "200" ]]; then
    echo "ERREUR: lien Taiga introuvable après webhook" >&2
    exit 1
  fi

  if ! grep -qE 'externalRef|ExternalRef|externalUrl|ExternalURL' "$BODY_FILE"; then
    echo "ERREUR: réponse GET sans ref ni URL" >&2
    exit 1
  fi

  echo "→ Lien Taiga OK pour demande ${DEMAND_ID}"
}

verify_link_get
echo "→ Fiche TMA : ${BASE_URL%/}/tma/${DEMAND_ID}"
