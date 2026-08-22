#!/usr/bin/env bash
# Route kore.ll-it-sc.be/api/v1/* vers kore-api (Cloud Run).
# Prérequis : setup-custom-domain.sh déjà exécuté.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

API_NEG="${KORE_API_NEG:-kore-api-neg}"
API_BACKEND="${KORE_API_BACKEND:-kore-api-backend}"
API_SERVICE="${API_SERVICE:-kore-api}"
FRONTEND_BACKEND="${KORE_FRONTEND_BACKEND:-kore-frontend-backend}"
SPLIT_MATCHER="${KORE_SPLIT_MATCHER:-kore-split}"
PATH_PREFIX="/api/v1/*"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== Kore API LB route — ${CUSTOM_DOMAIN}${PATH_PREFIX} ==="

echo "→ NEG serverless (${API_NEG})"
if ! gcloud compute network-endpoint-groups describe "$API_NEG" \
  --region="$GCP_RUN_REGION" --project="$GCP_PROJECT_ID" &>/dev/null; then
  gcloud compute network-endpoint-groups create "$API_NEG" \
    --region="$GCP_RUN_REGION" \
    --network-endpoint-type=serverless \
    --cloud-run-service="$API_SERVICE" \
    --project="$GCP_PROJECT_ID"
fi

echo "→ Backend service (${API_BACKEND})"
if ! gcloud compute backend-services describe "$API_BACKEND" \
  --global --project="$GCP_PROJECT_ID" &>/dev/null; then
  gcloud compute backend-services create "$API_BACKEND" \
    --global \
    --load-balancing-scheme=EXTERNAL \
    --project="$GCP_PROJECT_ID"
  gcloud compute backend-services add-backend "$API_BACKEND" \
    --global \
    --network-endpoint-group="$API_NEG" \
    --network-endpoint-group-region="$GCP_RUN_REGION" \
    --project="$GCP_PROJECT_ID"
fi

if gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
  --format='value(pathMatchers.name)' | tr ';' '\n' | grep -qx "$SPLIT_MATCHER"; then
  echo "→ Path matcher ${SPLIT_MATCHER} déjà présent"
else
  echo "→ Path matcher ${SPLIT_MATCHER} (/api/v1/* → ${API_BACKEND})"
  gcloud compute url-maps add-path-matcher "$URL_MAP" \
    --global --project="$GCP_PROJECT_ID" \
    --path-matcher-name="$SPLIT_MATCHER" \
    --default-service="$FRONTEND_BACKEND" \
    --path-rules="${PATH_PREFIX}=${API_BACKEND}"
fi

CURRENT_MATCHER="$(gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
  --format=json | python3 -c "
import json, sys
data = json.load(sys.stdin)
for rule in data.get('hostRules', []):
    if '${CUSTOM_DOMAIN}' in rule.get('hosts', []):
        print(rule.get('pathMatcher', ''))
        break
")"

if [[ "$CURRENT_MATCHER" == "$SPLIT_MATCHER" ]]; then
  echo "→ Hôte ${CUSTOM_DOMAIN} → ${SPLIT_MATCHER} (OK)"
else
  echo "→ Bascule hôte ${CUSTOM_DOMAIN} → ${SPLIT_MATCHER}"
  if [[ -n "$CURRENT_MATCHER" ]]; then
    gcloud compute url-maps remove-host-rule "$URL_MAP" \
      --global --project="$GCP_PROJECT_ID" \
      --host="$CUSTOM_DOMAIN" \
      --delete-orphaned-path-matcher 2>/dev/null \
      || gcloud compute url-maps remove-host-rule "$URL_MAP" \
        --global --project="$GCP_PROJECT_ID" \
        --host="$CUSTOM_DOMAIN"
  fi
  gcloud compute url-maps add-host-rule "$URL_MAP" \
    --global --project="$GCP_PROJECT_ID" \
    --hosts="$CUSTOM_DOMAIN" \
    --path-matcher-name="$SPLIT_MATCHER"
fi

if gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
  --format='json(hostRules)' | grep -q '"*"'; then
  echo "AVERTISSEMENT: règle hôte wildcard (*) détectée — à supprimer manuellement" >&2
fi

echo "→ Accès public Cloud Run (${API_SERVICE})"
gcloud run services add-iam-policy-binding "$API_SERVICE" \
  --region="$GCP_RUN_REGION" \
  --member="allUsers" \
  --role="roles/run.invoker" \
  --project="$GCP_PROJECT_ID" \
  --quiet 2>/dev/null || true

cat <<EOF

OK — tester :
  curl -sS -o /dev/null -w '%{http_code}\\n' \\
    -X POST https://${CUSTOM_DOMAIN}/api/v1/integrations/taiga/webhook \\
    -H 'Content-Type: application/json' -d '{}'
  (attendu : 401 invalid webhook secret, pas 302 HTML)

EOF
