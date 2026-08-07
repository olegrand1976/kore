#!/usr/bin/env bash
# Applique les grants kore_app sur Cloud SQL (idempotent).
# Source de vérité : internal/platform/db/grants_kore_app.sql
# Usage: ./infra/gcp/apply-db-grants.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

GRANTS_SQL="${REPO_ROOT}/internal/platform/db/grants_kore_app.sql"
if [[ ! -f "$GRANTS_SQL" ]]; then
  echo "ERREUR: grants introuvables: ${GRANTS_SQL}" >&2
  exit 1
fi

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

MIGRATE_PASS="$(gcloud secrets versions access latest --secret=kore-migrate-db-password --project="$GCP_PROJECT_ID")"
PROXY_PORT="${KORE_GRANTS_PROXY_PORT:-9484}"
cloud-sql-proxy "$CLOUDSQL_INSTANCE" --port "$PROXY_PORT" &
PROXY_PID=$!
trap 'kill "$PROXY_PID" 2>/dev/null || true' EXIT
sleep 3

echo "→ Grants kore_app (${GRANTS_SQL})"
PGPASSWORD="$MIGRATE_PASS" psql -h 127.0.0.1 -p "$PROXY_PORT" -U "$MIGRATE_USER" -d "$DB_NAME" \
  -v ON_ERROR_STOP=1 -f "$GRANTS_SQL"
echo "→ Grants kore_app OK"
