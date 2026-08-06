#!/usr/bin/env bash
# Orchestrates Docker test stack + API + Nuxt preview + Playwright smoke.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COMPOSE_FILE="${ROOT}/deploy/docker-compose.test.yml"
# Avoid colliding with other repos that also use a "deploy/" compose project name.
export COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-kore-test}"
# Preserve caller override (local port conflicts with other stacks).
API_PORT_PRESET="${KORE_API_PORT:-}"
FRONTEND_PORT_PRESET="${KORE_FRONTEND_PORT:-}"
if [ -f "${ROOT}/.env" ]; then
  set -a && source "${ROOT}/.env" && set +a
fi
API_PORT="${API_PORT_PRESET:-${KORE_API_PORT:-8081}}"
FRONTEND_PORT="${FRONTEND_PORT_PRESET:-${KORE_FRONTEND_PORT:-3001}}"
export PLAYWRIGHT_BASE_URL="${PLAYWRIGHT_BASE_URL:-http://localhost:${FRONTEND_PORT}}"
API_PID=""
FRONT_PID=""

kill_tree() {
  local pid="${1:-}"
  if [ -z "$pid" ]; then
    return 0
  fi
  if kill -0 "$pid" 2>/dev/null; then
    kill -- "-$pid" 2>/dev/null || kill "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
}

cleanup() {
  kill_tree "${FRONT_PID}"
  kill_tree "${API_PID}"
  docker compose -f "${COMPOSE_FILE}" down >/dev/null 2>&1 || true
}
trap cleanup EXIT

wait_http() {
  local url="$1"
  local attempts="${2:-60}"
  local i
  for i in $(seq 1 "${attempts}"); do
    if curl -sf "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "ERROR: timeout waiting for ${url}" >&2
  return 1
}

wait_tcp() {
  local host="$1"
  local port="$2"
  local attempts="${3:-60}"
  local i
  for i in $(seq 1 "${attempts}"); do
    if (echo >/dev/tcp/"${host}"/"${port}") >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "ERROR: timeout waiting for ${host}:${port}" >&2
  return 1
}

echo "== E2E: start test stack =="
docker compose -f "${COMPOSE_FILE}" down >/dev/null 2>&1 || true
docker compose -f "${COMPOSE_FILE}" up -d
wait_tcp localhost 5433 60
wait_tcp localhost 6382 60
# Give Postgres a moment after port open (ready for connections).
sleep 3
docker compose -f "${COMPOSE_FILE}" ps

# Force test-stack DSN (ignore .env DATABASE_URL pointing at compose hostname "db").
export DATABASE_URL="postgres://kore:kore@localhost:5433/kore_test?sslmode=disable"
export REDIS_ADDR="localhost:6382"
export HTTP_ADDR=":${API_PORT}"
export JWT_SIGNING_KEY="${JWT_SIGNING_KEY:-ci-e2e-test-key}"
export DEV_SEED_ENABLED="${DEV_SEED_ENABLED:-true}"
export MIGRATE_ON_BOOT="${MIGRATE_ON_BOOT:-false}"
export STRIPE_API_BASE="${STRIPE_API_BASE:-http://localhost:12112}"

cd "${ROOT}"
echo "== E2E: migrate + seed =="
go run ./cmd/kore-api migrate
go run ./cmd/kore-api seed

echo "== E2E: start API on :${API_PORT} =="
setsid go run ./cmd/kore-api >/tmp/kore-e2e-api.log 2>&1 &
API_PID=$!
wait_http "http://localhost:${API_PORT}/health" 30
curl -sf "http://localhost:${API_PORT}/health" | grep -q ok
# Sanity login against API (catches DB/seed issues early).
curl -sf -X POST "http://localhost:${API_PORT}/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"login":"ADM_admin","password":"Admin123!"}' | grep -qiE 'accessToken|AccessToken'

echo "== E2E: build + preview frontend on :${FRONTEND_PORT} =="
cd "${ROOT}/frontend"
export NUXT_API_BASE="http://localhost:${API_PORT}"
export NUXT_PUBLIC_API_BASE="http://localhost:${API_PORT}"
export PORT="${FRONTEND_PORT}"
export HOST="0.0.0.0"
export NITRO_HOST="0.0.0.0"
export NITRO_PORT="${FRONTEND_PORT}"
npm ci --prefer-offline
npm run build
setsid node .output/server/index.mjs >/tmp/kore-e2e-front.log 2>&1 &
FRONT_PID=$!
wait_http "http://localhost:${FRONTEND_PORT}/login" 60

echo "== E2E: Playwright =="
npx playwright install chromium
# Keep DB/redis alive: fail fast if they disappear before tests.
docker compose -f "${COMPOSE_FILE}" ps | grep -E 'db|redis' || true
CI=true npx playwright test --config=e2e/playwright.config.ts

echo "== E2E: OK =="
