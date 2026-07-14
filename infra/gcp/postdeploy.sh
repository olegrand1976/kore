#!/usr/bin/env bash
# Postdeploy Kore : jobs + smoke (+ migrate/seed optionnels).
# Usage: ./infra/gcp/postdeploy.sh [--migrate] [--seed | --seed-reset]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

RUN_MIGRATE=false
RUN_SEED=false
RUN_SEED_RESET=false
for arg in "$@"; do
  case "$arg" in
    --migrate) RUN_MIGRATE=true ;;
    --seed) RUN_SEED=true ;;
    --seed-reset) RUN_SEED_RESET=true ;;
    --skip-seed) RUN_SEED=false; RUN_SEED_RESET=false ;;
  esac
done

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

run_job() {
  local name="$1"
  echo "→ Exécution ${name}"
  gcloud run jobs execute "$name" \
    --region="$GCP_RUN_REGION" --project="$GCP_PROJECT_ID" \
    --wait --quiet
}

echo "=== Kore postdeploy — ${GCP_PROJECT_ID} ==="

bash "${SCRIPT_DIR}/setup-jobs.sh"

if $RUN_MIGRATE; then
  run_job "kore-migrate"
fi

if $RUN_SEED_RESET; then
  run_job "kore-seed-reset"
elif $RUN_SEED; then
  run_job "kore-seed"
fi

bash "${SCRIPT_DIR}/smoke-test.sh"

API_URL="$(api_run_url)"
FE_URL="$(frontend_run_url)"

cat <<EOF

Postdeploy terminé.
  API       : ${API_URL:-non déployée}
  Frontend  : ${FE_URL:-non déployé}
  Custom    : ${PUBLIC_SITE_URL}
  Admin     : ADM_admin (mot de passe seed — voir internal/seed/constants.go)

EOF
