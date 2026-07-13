#!/usr/bin/env bash
# Domaine custom kore.ll-it-sc.be via Load Balancer global + Serverless NEG.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

DOMAIN="${CUSTOM_DOMAIN:-kore.ll-it-sc.be}"
SERVICE="${FRONTEND_SERVICE:-kore-frontend}"
NEG_NAME="${NEG_NAME:-kore-frontend-neg}"
BACKEND_NAME="${BACKEND_NAME:-kore-frontend-backend}"
PATH_MATCHER="${PATH_MATCHER:-kore-frontend}"
CERT_NAME="${CERT_NAME:-kore-ll-it-sc-cert}"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

proxy_cert_names() {
  gcloud compute target-https-proxies describe "$HTTPS_PROXY" \
    --global --project="$GCP_PROJECT_ID" \
    --format='json(sslCertificates)' \
    | python3 -c "import json,sys; data=json.load(sys.stdin); print(','.join(u.rsplit('/',1)[-1] for u in data))"
}

renew_managed_cert() {
  local domain_status cert_status existing
  cert_status="$(gcloud compute ssl-certificates describe "$CERT_NAME" \
    --global --project="$GCP_PROJECT_ID" \
    --format='value(managed.status)' 2>/dev/null || true)"
  domain_status="$(gcloud compute ssl-certificates describe "$CERT_NAME" \
    --global --project="$GCP_PROJECT_ID" \
    --format="value(managed.domainStatus['${DOMAIN}'])" 2>/dev/null || true)"

  if [[ "$cert_status" == "ACTIVE" && "$domain_status" == "ACTIVE" ]]; then
    echo "  Certificat ${CERT_NAME} déjà ACTIVE"
    return 0
  fi
  if [[ "$domain_status" != "FAILED_NOT_VISIBLE" && "$domain_status" != "PROVISIONING_FAILED" && "$cert_status" != "PROVISIONING_FAILED" ]]; then
    echo "  Certificat ${CERT_NAME} : ${cert_status} / ${DOMAIN}=${domain_status:-pending}"
    return 0
  fi

  echo "→ Renouvellement certificat (${domain_status:-${cert_status}}) — DNS doit pointer vers ${LB_IP}"
  existing="$(proxy_cert_names | tr ',' '\n' | grep -vx "$CERT_NAME" | paste -sd, -)"
  gcloud compute target-https-proxies update "$HTTPS_PROXY" \
    --global --project="$GCP_PROJECT_ID" \
    --ssl-certificates="$existing" --quiet
  gcloud compute ssl-certificates delete "$CERT_NAME" \
    --global --project="$GCP_PROJECT_ID" --quiet
  gcloud compute ssl-certificates create "$CERT_NAME" \
    --domains="$DOMAIN" --global --project="$GCP_PROJECT_ID"
  if [[ -n "$existing" ]]; then
    gcloud compute target-https-proxies update "$HTTPS_PROXY" \
      --global --project="$GCP_PROJECT_ID" \
      --ssl-certificates="${existing},${CERT_NAME}" --quiet
  else
    gcloud compute target-https-proxies update "$HTTPS_PROXY" \
      --global --project="$GCP_PROJECT_ID" \
      --ssl-certificates="${CERT_NAME}" --quiet
  fi
}

echo "→ NEG serverless (${NEG_NAME})"
if ! gcloud compute network-endpoint-groups describe "$NEG_NAME" \
  --region="$GCP_RUN_REGION" --project="$GCP_PROJECT_ID" &>/dev/null; then
  gcloud compute network-endpoint-groups create "$NEG_NAME" \
    --region="$GCP_RUN_REGION" \
    --network-endpoint-type=serverless \
    --cloud-run-service="$SERVICE" \
    --project="$GCP_PROJECT_ID"
fi

echo "→ Backend service (${BACKEND_NAME})"
if ! gcloud compute backend-services describe "$BACKEND_NAME" \
  --global --project="$GCP_PROJECT_ID" &>/dev/null; then
  gcloud compute backend-services create "$BACKEND_NAME" \
    --global \
    --load-balancing-scheme=EXTERNAL \
    --project="$GCP_PROJECT_ID"
  gcloud compute backend-services add-backend "$BACKEND_NAME" \
    --global \
    --network-endpoint-group="$NEG_NAME" \
    --network-endpoint-group-region="$GCP_RUN_REGION" \
    --project="$GCP_PROJECT_ID"
fi

echo "→ Règle hôte sur ${URL_MAP}"
if gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
  --format='json(hostRules)' | grep -q '"\*"'; then
  echo "→ Suppression règle hôte wildcard erronée (* → ${PATH_MATCHER})"
  gcloud compute url-maps remove-host-rule "$URL_MAP" \
    --global --project="$GCP_PROJECT_ID" --host='*' 2>/dev/null \
    || echo "  (wildcard * absent ou déjà retiré)"
fi
if ! gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
  --format='value(pathMatchers.name)' | tr ';' '\n' | grep -qx "$PATH_MATCHER"; then
  gcloud compute url-maps add-path-matcher "$URL_MAP" \
    --global \
    --path-matcher-name="$PATH_MATCHER" \
    --default-service="$BACKEND_NAME" \
    --project="$GCP_PROJECT_ID"
fi
if ! gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
  --format='json(hostRules)' | grep -q "\"${DOMAIN}\""; then
  gcloud compute url-maps add-host-rule "$URL_MAP" \
    --global \
    --hosts="$DOMAIN" \
    --path-matcher-name="$PATH_MATCHER" \
    --project="$GCP_PROJECT_ID"
else
  echo "  Règle hôte ${DOMAIN} déjà présente"
fi

echo "→ Certificat managé (${CERT_NAME})"
if ! gcloud compute ssl-certificates describe "$CERT_NAME" \
  --global --project="$GCP_PROJECT_ID" &>/dev/null; then
  gcloud compute ssl-certificates create "$CERT_NAME" \
    --domains="$DOMAIN" \
    --global \
    --project="$GCP_PROJECT_ID"
else
  renew_managed_cert
fi

EXISTING_CERTS="$(proxy_cert_names)"
if ! echo ",${EXISTING_CERTS}," | grep -q ",${CERT_NAME},"; then
  gcloud compute target-https-proxies update "$HTTPS_PROXY" \
    --global \
    --ssl-certificates="${EXISTING_CERTS},${CERT_NAME}" \
    --project="$GCP_PROJECT_ID"
fi

echo "→ Accès public Cloud Run (${SERVICE})"
gcloud run services add-iam-policy-binding "$SERVICE" \
  --region="$GCP_RUN_REGION" \
  --member="allUsers" \
  --role="roles/run.invoker" \
  --project="$GCP_PROJECT_ID" \
  --quiet 2>/dev/null || true

cat <<EOF

OK — prochaines étapes manuelles (OVH, zone ll-it-sc.be) :

1. Créer un enregistrement A :
   - Sous-domaine : kore
   - Cible        : ${LB_IP}
2. Attendre propagation DNS + certificat ACTIVE :
   gcloud compute ssl-certificates describe ${CERT_NAME} --global --format='yaml(managed)'
3. Tester :
   curl -I https://${DOMAIN}/

EOF
