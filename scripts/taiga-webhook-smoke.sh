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
SMOKE_TOKEN=""

jwt_tenant_id() {
  local token="$1"
  python3 -c "import base64,json,sys; t=sys.argv[1].split('.')[1]; t+='='*(-len(t)%4); print(json.loads(base64.urlsafe_b64decode(t)).get('tenant_id',''))" "$token"
}

resolve_demand_id() {
  local login="${SMOKE_LOGIN:-ADM_admin}"
  local password="${SMOKE_PASSWORD:-Admin123!}"
  local login_resp token demands

  if [[ -n "$DEMAND_ID" && -n "${SMOKE_TOKEN:-}" ]]; then
    return 0
  fi

  echo "→ Auth API (${login})…"
  if ! login_resp=$(curl -sS -f -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"login\":\"${login}\",\"password\":\"${password}\"}" 2>/dev/null); then
    if [[ -z "$DEMAND_ID" ]]; then
      echo "ERREUR: login impossible — passez DEMAND_UUID en 2e argument" >&2
      exit 1
    fi
    return 0
  fi
  SMOKE_TOKEN=$(echo "$login_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('accessToken') or d.get('AccessToken') or '')")
  if [[ -z "$SMOKE_TOKEN" ]]; then
    echo "ERREUR: token absent" >&2
    exit 1
  fi
  TENANT="$(jwt_tenant_id "$SMOKE_TOKEN")"
  if [[ -z "$TENANT" ]]; then
    TENANT="${TAIGA_DEFAULT_TENANT_ID:-00000000-0000-4000-8000-000000000001}"
  fi

  if [[ -n "$DEMAND_ID" ]]; then
    return 0
  fi

  echo "→ Résolution demand_id (première demande TMA, tenant ${TENANT})…"
  if ! demands=$(curl -sS -f -H "Authorization: Bearer ${SMOKE_TOKEN}" "${API_BASE}/demands?page_size=1" 2>/dev/null); then
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

TENANT="${TENANT:-${TAIGA_DEFAULT_TENANT_ID:-00000000-0000-4000-8000-000000000001}}"
WEBHOOK_URL="${API_BASE}/integrations/taiga/webhook"
SMOKE_STORY_ID="${TAIGA_SMOKE_STORY_ID:-99942}"

payload=$(cat <<EOF
{
  "action": "create",
  "type": "userstory",
  "data": {
    "id": ${SMOKE_STORY_ID},
    "ref": 42,
    "project": { "slug": "kore-tma" },
    "external_reference": ["kore", "${DEMAND_ID}"],
    "permalink": "https://taiga.ll-it-sc.be/project/kore-tma/us/42"
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

  token="${SMOKE_TOKEN:-}"
  if [[ -z "$token" ]]; then
    echo "→ POST ${API_BASE}/auth/login (${login})"
    if ! login_resp=$(curl -sS -f -X POST "${API_BASE}/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"login\":\"${login}\",\"password\":\"${password}\"}" 2>/dev/null); then
      echo "→ GET lien ignoré (login impossible — définir SMOKE_LOGIN/SMOKE_PASSWORD)" >&2
      return 0
    fi
    token=$(echo "$login_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('accessToken') or d.get('AccessToken') or '')")
  fi
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
