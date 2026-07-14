#!/usr/bin/env bash
# Applique la config Google OIDC (Secret Manager + base org.identity_providers).
# Usage:
#   export OIDC_GOOGLE_CLIENT_ID=...
#   export OIDC_GOOGLE_CLIENT_SECRET=...
#   ./infra/gcp/apply-oidc-config.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

if [[ -f "${REPO_ROOT}/.env.oidc" ]]; then
  # shellcheck disable=SC1091
  set -a && source "${REPO_ROOT}/.env.oidc" && set +a
elif [[ -f "${REPO_ROOT}/.env.iodc" ]]; then
  # Compat: certains environnements ont un fichier mal nommé ".env.iodc"
  # shellcheck disable=SC1091
  set -a && source "${REPO_ROOT}/.env.iodc" && set +a
fi

CLIENT_ID="${OIDC_GOOGLE_CLIENT_ID:-}"
CLIENT_SECRET="${OIDC_GOOGLE_CLIENT_SECRET:-}"

if [[ -z "$CLIENT_ID" || -z "$CLIENT_SECRET" ]]; then
  echo "ERREUR: renseignez OIDC_GOOGLE_CLIENT_ID et OIDC_GOOGLE_CLIENT_SECRET (.env.oidc ou env)" >&2
  exit 1
fi

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

add_secret_version() {
  local name="$1"
  local value="$2"
  if ! gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    gcloud secrets create "$name" --replication-policy=automatic --project="$GCP_PROJECT_ID" --quiet
  fi
  echo -n "$value" | gcloud secrets versions add "$name" --data-file=- --project="$GCP_PROJECT_ID" --quiet
}

add_secret_version "kore-oidc-google-client-id" "$CLIENT_ID"
add_secret_version "kore-oidc-google-client-secret" "$CLIENT_SECRET"
echo "→ Secrets kore-oidc-google-client-* mis à jour"

MIGRATE_PASS="$(gcloud secrets versions access latest --secret=kore-migrate-db-password --project="$GCP_PROJECT_ID")"
PROXY_PORT=9472
cloud-sql-proxy "$CLOUDSQL_INSTANCE" --port "$PROXY_PORT" &
PROXY_PID=$!
trap 'kill "$PROXY_PID" 2>/dev/null || true' EXIT
sleep 3

PGPASSWORD="$MIGRATE_PASS" psql -h 127.0.0.1 -p "$PROXY_PORT" -U "$MIGRATE_USER" -d "$DB_NAME" <<SQL
INSERT INTO org.identity_providers (
  id, tenant_id, name, issuer, client_id, client_secret, jwks_uri, scopes, default_profile, enabled
) VALUES (
  '00000000-0000-4000-8000-000000000020',
  '00000000-0000-4000-8000-000000000001',
  'Google',
  'https://accounts.google.com',
  '${CLIENT_ID}',
  '${CLIENT_SECRET}',
  'https://www.googleapis.com/oauth2/v3/certs',
  'openid profile email',
  'Collaborateur',
  TRUE
)
ON CONFLICT (tenant_id) DO UPDATE SET
  name = EXCLUDED.name,
  issuer = EXCLUDED.issuer,
  client_id = EXCLUDED.client_id,
  client_secret = EXCLUDED.client_secret,
  jwks_uri = EXCLUDED.jwks_uri,
  scopes = EXCLUDED.scopes,
  enabled = TRUE,
  updated_at = NOW();
SQL

echo "→ IdP Google activé sur ${DB_NAME} (${GCP_PROJECT_ID})"
echo "Redirect URIs à autoriser dans Google Cloud Console :"
echo "  http://localhost:3001/login"
echo "  https://${CUSTOM_DOMAIN}/login"
