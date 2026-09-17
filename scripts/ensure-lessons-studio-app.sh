#!/usr/bin/env bash
# Ensure application "Lessons-studio" exists on the ADM_olivier tenant and create
# an API key for Lessons Studio open tickets (POST /api/v1/open/tickets).
#
# Usage:
#   KORE_API_URL=https://kore.ll-it-sc.be/api/v1 \
#   KORE_LOGIN=ADM_olivier \
#   KORE_PASSWORD='…' \
#   ./scripts/ensure-lessons-studio-app.sh
#
# Prints env vars for create-courses:
#   KORE_API_URL, KORE_API_KEY, KORE_APPLICATION_ID, KORE_APPLICATION_LIBELLE
set -euo pipefail

API_BASE="${KORE_API_URL:-http://127.0.0.1:8081/api/v1}"
API_BASE="${API_BASE%/}"
LOGIN="${KORE_LOGIN:-ADM_olivier}"
PASSWORD="${KORE_PASSWORD:-}"
LIBELLE="${KORE_APPLICATION_LIBELLE:-Lessons-studio}"
KEY_NAME="${KORE_API_KEY_NAME:-lessons-studio}"

if [[ -z "$PASSWORD" ]]; then
  echo "ERREUR: définir KORE_PASSWORD (compte ${LOGIN})" >&2
  exit 1
fi

echo "→ Login ${LOGIN} @ ${API_BASE}…"
login_resp=$(curl -sS -f -X POST "${API_BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"login\":\"${LOGIN}\",\"password\":\"${PASSWORD}\"}")
TOKEN=$(echo "$login_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('accessToken') or d.get('AccessToken') or '')")
if [[ -z "$TOKEN" ]]; then
  echo "ERREUR: accessToken absent" >&2
  echo "$login_resp" >&2
  exit 1
fi

echo "→ List applications…"
apps_resp=$(curl -sS -f -H "Authorization: Bearer ${TOKEN}" "${API_BASE}/applications?active=all")
APP_ID=$(echo "$apps_resp" | LIBELLE="$LIBELLE" python3 -c '
import json, os, re, sys
libelle = os.environ["LIBELLE"].strip()
want = libelle.lower()
want_norm = re.sub(r"[^a-z0-9]", "", want)
aliases = {"lessons-studio", "lesson-studio", "lesson studio", "lessons studio"}
doc = json.load(sys.stdin)
items = doc.get("data") or doc
if isinstance(items, dict):
    items = items.get("items") or items.get("applications") or []
app_id = ""
resolved = ""
for app in items or []:
    got = str(app.get("libelle") or "").strip()
    got_l = got.lower()
    got_norm = re.sub(r"[^a-z0-9]", "", got_l)
    if got_l == want or got_norm == want_norm or (want in aliases and got_l in aliases):
        app_id = str(app.get("id") or "")
        resolved = got
        break
print(app_id)
if resolved:
    print(resolved, file=sys.stderr)
')

if [[ -z "$APP_ID" ]]; then
  echo "→ Resolve a site share for create…"
  sites_resp=$(curl -sS -f -H "Authorization: Bearer ${TOKEN}" "${API_BASE}/sites")
  SITE_ID=$(echo "$sites_resp" | python3 -c '
import json, sys
doc = json.load(sys.stdin)
items = doc.get("data") or doc
if isinstance(items, dict):
    items = items.get("items") or items.get("sites") or []
print(str((items or [{}])[0].get("id") or "") if items else "")
')
  if [[ -z "$SITE_ID" ]]; then
    echo "ERREUR: aucun site sur le tenant — impossible de créer l'application (share requis)." >&2
    exit 1
  fi
  echo "→ Create application ${LIBELLE} (site ${SITE_ID})…"
  create_resp=$(curl -sS -f -X POST "${API_BASE}/applications" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"libelle\":\"${LIBELLE}\",\"proprietaire\":\"Lessons Studio\",\"modeFacturation\":\"non\",\"methodologyProfile\":\"psa\",\"uoActivee\":false,\"siteIds\":[\"${SITE_ID}\"]}")
  APP_ID=$(echo "$create_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('id') or '')")
  if [[ -z "$APP_ID" ]]; then
    echo "ERREUR: création application échouée" >&2
    echo "$create_resp" >&2
    exit 1
  fi
  echo "  créée: ${APP_ID}"
else
  echo "  déjà présente: ${APP_ID}"
fi

echo "→ Create API key ${KEY_NAME}…"
key_resp=$(curl -sS -X POST "${API_BASE}/integrations/api-keys" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${KEY_NAME}\"}")
PLAIN_KEY=$(echo "$key_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('plainKey') or d.get('PlainKey') or '')")
if [[ -z "$PLAIN_KEY" ]]; then
  # Name collision: create a unique key name.
  KEY_NAME="${KEY_NAME}-$(date +%s)"
  echo "  retry with name ${KEY_NAME}…"
  key_resp=$(curl -sS -f -X POST "${API_BASE}/integrations/api-keys" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"${KEY_NAME}\"}")
  PLAIN_KEY=$(echo "$key_resp" | python3 -c "import json,sys; d=json.load(sys.stdin).get('data') or {}; print(d.get('plainKey') or d.get('PlainKey') or '')")
fi
if [[ -z "$PLAIN_KEY" ]]; then
  echo "ERREUR: plainKey absent" >&2
  echo "$key_resp" >&2
  exit 1
fi

echo
echo "# --- add to create-courses deploy/.env ---"
echo "KORE_API_URL=${API_BASE}"
echo "KORE_API_KEY=${PLAIN_KEY}"
echo "KORE_APPLICATION_ID=${APP_ID}"
echo "KORE_APPLICATION_LIBELLE=${LIBELLE}"
echo
echo "Smoke: curl -sS -X POST ${API_BASE}/open/tickets -H \"X-Api-Key: \$KORE_API_KEY\" -H 'Content-Type: application/json' -d '{\"subject\":\"ping\",\"description\":\"bootstrap smoke\"}'"
