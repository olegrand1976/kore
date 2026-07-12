#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if [ -f "$ROOT/.env" ]; then
  set -a && source "$ROOT/.env" && set +a
fi
API_PORT="${KORE_API_PORT:-8081}"

echo "== Kore MVP smoke test (API :$API_PORT) =="

curl -sf "http://localhost:${API_PORT}/health" | grep -q ok
curl -sf "http://localhost:${API_PORT}/ready" | grep -q ready

LOGIN_RESP=$(curl -sf -X POST "http://localhost:${API_PORT}/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"login":"ADM_admin","password":"Admin123!"}')

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.AccessToken // .data.accessToken')
REFRESH=$(echo "$LOGIN_RESP" | jq -r '.data.RefreshToken // .data.refreshToken')

test -n "$TOKEN"
test "$TOKEN" != "null"

curl -sf "http://localhost:${API_PORT}/api/v1/societes" -H "Authorization: Bearer $TOKEN" >/dev/null
curl -sf "http://localhost:${API_PORT}/api/v1/public/pricing" >/dev/null
curl -sf "http://localhost:${API_PORT}/api/v1/public/modules" >/dev/null

# Auth refresh
if [ -n "$REFRESH" ] && [ "$REFRESH" != "null" ]; then
  curl -sf -X POST "http://localhost:${API_PORT}/api/v1/auth/refresh" \
    -H 'Content-Type: application/json' \
    -d "{\"refreshToken\":\"$REFRESH\"}" >/dev/null
fi

# CRA flow: get or create current month timesheet
MONTH=$(date +%Y-%m)
CRA=$(curl -sf "http://localhost:${API_PORT}/api/v1/timesheets?month=$MONTH" \
  -H "Authorization: Bearer $TOKEN")
CRA_ID=$(echo "$CRA" | jq -r '.data.id // .data.ID')
test -n "$CRA_ID"
test "$CRA_ID" != "null"

curl -sf -X PUT "http://localhost:${API_PORT}/api/v1/timesheets/${CRA_ID}/weeks/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"lines":[{"sourceType":"manual","sourceId":"default","day":"'"$MONTH"'-01","duration":240,"comment":"smoke"}]}' >/dev/null

curl -sf -X POST "http://localhost:${API_PORT}/api/v1/timesheets/${CRA_ID}/weeks/1/submit" \
  -H "Authorization: Bearer $TOKEN" >/dev/null

# Billing checkout (stripe-mock)
CHECKOUT=$(curl -sf -X POST "http://localhost:${API_PORT}/api/v1/billing/checkout-session" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"modules":["org","cra"],"seats":5,"successUrl":"http://localhost:3001/billing/success","cancelUrl":"http://localhost:3001/billing/cancel"}')
CHECKOUT_URL=$(echo "$CHECKOUT" | jq -r '.data.url // .data.URL')
test -n "$CHECKOUT_URL"
test "$CHECKOUT_URL" != "null"

# Public booking slots
curl -sf "http://localhost:${API_PORT}/api/v1/public/booking/slots" >/dev/null

echo "Smoke test OK"
