#!/usr/bin/env bash
# Smoke test Kore déployé sur GCP (health, ready, frontend).
# Usage: ./infra/gcp/smoke-test.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

API_URL="${KORE_API_URL:-$(api_run_url)}"
FE_URL="${KORE_FRONTEND_URL:-$(frontend_run_url)}"
CUSTOM_URL="${KORE_CUSTOM_URL:-$PUBLIC_SITE_URL}"

fail() {
  echo "ERREUR: $*" >&2
  exit 1
}

check_url() {
  local label="$1"
  local url="$2"
  local pattern="${3:-}"
  [[ -n "$url" ]] || fail "${label} : URL vide"
  echo "→ ${label} : ${url}"
  if [[ -n "$pattern" ]]; then
    curl -sf "$url" | grep -q "$pattern" || fail "${label} : pattern '${pattern}' introuvable"
  else
    curl -sfI "$url" >/dev/null || fail "${label} : HTTP indisponible"
  fi
  echo "  OK"
}

echo "=== Kore smoke test GCP ==="

[[ -n "$API_URL" ]] || fail "Service ${API_SERVICE} introuvable — déployez d'abord"
check_url "API /health" "${API_URL}/health" "ok"
check_url "API /ready" "${API_URL}/ready" "ready"

if [[ -n "$FE_URL" ]]; then
  check_url "Frontend Cloud Run" "$FE_URL"
fi

if [[ -n "$CUSTOM_URL" && "$CUSTOM_URL" != "$FE_URL" ]]; then
  if curl -sfI "$CUSTOM_URL" >/dev/null 2>&1; then
    check_url "Domaine custom" "$CUSTOM_URL"
  else
    echo "→ Domaine custom ${CUSTOM_URL} pas encore joignable (DNS/LB en attente)"
  fi
fi

echo ""
echo "Smoke test GCP OK"
