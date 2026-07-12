#!/usr/bin/env bash
set -euo pipefail

echo "== Kore MVP smoke test =="

curl -sf http://localhost:8080/health | grep -q ok
curl -sf http://localhost:8080/ready | grep -q ready

TOKEN=$(curl -sf -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"login":"ADM_admin","password":"Admin123!"}' | jq -r '.data.accessToken')

test -n "$TOKEN"

curl -sf http://localhost:8080/api/v1/societes -H "Authorization: Bearer $TOKEN" >/dev/null
curl -sf http://localhost:8080/api/v1/public/pricing >/dev/null
curl -sf http://localhost:8080/api/v1/public/modules >/dev/null

echo "Smoke test OK"
