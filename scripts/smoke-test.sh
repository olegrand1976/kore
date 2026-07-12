#!/usr/bin/env bash
set -euo pipefail

# Charge les ports depuis .env (racine du projet)
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if [ -f "$ROOT/.env" ]; then
  # shellcheck disable=SC1091
  set -a && source "$ROOT/.env" && set +a
fi
API_PORT="${KORE_API_PORT:-8081}"

echo "== Kore MVP smoke test (API :$API_PORT) =="

curl -sf "http://localhost:${API_PORT}/health" | grep -q ok
curl -sf "http://localhost:${API_PORT}/ready" | grep -q ready

TOKEN=$(curl -sf -X POST "http://localhost:${API_PORT}/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"login":"ADM_admin","password":"Admin123!"}' | jq -r '.data.AccessToken // .data.accessToken')

test -n "$TOKEN"
test "$TOKEN" != "null"

curl -sf "http://localhost:${API_PORT}/api/v1/societes" -H "Authorization: Bearer $TOKEN" >/dev/null
curl -sf "http://localhost:${API_PORT}/api/v1/public/pricing" >/dev/null
curl -sf "http://localhost:${API_PORT}/api/v1/public/modules" >/dev/null

echo "Smoke test OK"
