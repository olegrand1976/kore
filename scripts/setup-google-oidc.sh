#!/usr/bin/env bash
# Configure Google OIDC pour Kore (local + GCP).
# Prérequis : OIDC_GOOGLE_CLIENT_ID et OIDC_GOOGLE_CLIENT_SECRET dans .env.oidc
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_OIDC="${ROOT}/.env.oidc"
ENV_FILE="${ROOT}/.env"

CLIENT_ID="${OIDC_GOOGLE_CLIENT_ID:-}"
CLIENT_SECRET="${OIDC_GOOGLE_CLIENT_SECRET:-}"

if [[ -f "$ENV_OIDC" ]]; then
  # shellcheck disable=SC1091
  set -a && source "$ENV_OIDC" && set +a
  CLIENT_ID="${OIDC_GOOGLE_CLIENT_ID:-$CLIENT_ID}"
  CLIENT_SECRET="${OIDC_GOOGLE_CLIENT_SECRET:-$CLIENT_SECRET}"
fi

if [[ -z "$CLIENT_ID" || -z "$CLIENT_SECRET" ]]; then
  echo "ERREUR: OIDC_GOOGLE_CLIENT_ID et OIDC_GOOGLE_CLIENT_SECRET requis dans .env.oidc" >&2
  echo "Console : https://console.cloud.google.com/apis/credentials?project=premedica-prod-2025" >&2
  exit 1
fi

upsert_env() {
  local file="$1" key="$2" val="$3"
  if grep -q "^${key}=" "$file" 2>/dev/null; then
    sed -i "s|^${key}=.*|${key}=${val}|" "$file"
  else
    echo "${key}=${val}" >> "$file"
  fi
}

upsert_env "$ENV_OIDC" OIDC_GOOGLE_CLIENT_ID "$CLIENT_ID"
upsert_env "$ENV_OIDC" OIDC_GOOGLE_CLIENT_SECRET "$CLIENT_SECRET"
upsert_env "$ENV_FILE" OIDC_GOOGLE_CLIENT_ID "$CLIENT_ID"
upsert_env "$ENV_FILE" OIDC_GOOGLE_CLIENT_SECRET "$CLIENT_SECRET"

echo "→ .env.oidc et .env mis à jour"

KORE_DB_PORT="${KORE_DB_PORT:-5434}"
if [[ -f "$ENV_FILE" ]]; then
  KORE_DB_PORT="$(grep -E '^KORE_DB_PORT=' "$ENV_FILE" | cut -d= -f2 | tr -d '\r' || true)"
fi
KORE_DB_PORT="${KORE_DB_PORT:-5434}"

PGPASSWORD=kore psql -h localhost -p "$KORE_DB_PORT" -U kore -d kore <<SQL
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

echo "→ IdP Google activé en local"

if [[ -x "${ROOT}/infra/gcp/apply-oidc-config.sh" ]]; then
  export OIDC_GOOGLE_CLIENT_ID="$CLIENT_ID"
  export OIDC_GOOGLE_CLIENT_SECRET="$CLIENT_SECRET"
  bash "${ROOT}/infra/gcp/apply-oidc-config.sh"
fi

echo "✓ SSO Google configuré (local + GCP)"
