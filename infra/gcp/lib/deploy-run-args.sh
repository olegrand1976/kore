# Variables Cloud Run partagées (API, frontend, jobs). Source : infra/gcp/lib/gcp-env.sh
# shellcheck shell=bash
set -euo pipefail

kore_resolve_redis_addr() {
  local redis_url host port
  redis_url="$(gcloud secrets versions access latest \
    --secret=kore-redis-url --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -n "$redis_url" ]]; then
    host="$(python3 -c "from urllib.parse import urlparse; u=urlparse('$redis_url'); print(u.hostname or '')")"
    port="$(python3 -c "from urllib.parse import urlparse; u=urlparse('$redis_url'); print(u.port or 6379)")"
    if [[ -n "$host" ]]; then
      printf '%s:%s' "$host" "$port"
      return 0
    fi
  fi
  local vm_host
  vm_host="$(gcloud secrets versions access latest \
    --secret=premedica-redis-host --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$vm_host" ]]; then
    vm_host="$(gcloud compute instances describe "$REDIS_VM_NAME" \
      --zone="$REDIS_VM_ZONE" --project="$GCP_PROJECT_ID" \
      --format='value(networkInterfaces[0].networkIP)' 2>/dev/null || true)"
  fi
  printf '%s:6379' "${vm_host:-localhost}"
}

kore_write_api_env_file() {
  local path="$1"
  local redis_addr
  redis_addr="$(kore_resolve_redis_addr)"
  cat >"$path" <<EOF
HTTP_ADDR: ":8080"
LOG_LEVEL: "info"
MIGRATE_ON_BOOT: "false"
DEV_SEED_ENABLED: "false"
REDIS_ADDR: "${redis_addr}"
REDIS_DB: "${REDIS_DB}"
REDIS_KEY_PREFIX: "${REDIS_KEY_PREFIX}"
REDIS_TLS: "false"
JWT_TTL: "15m"
JWT_REFRESH_TTL: "168h"
BILLING_TRIAL_DAYS: "14"
SMTP_HOST: "pro1.mail.ovh.net"
SMTP_PORT: "587"
SMTP_FROM: "Kore <noreply@ll-it-sc.be>"
AI_LLM_PROVIDER: "gemini"
GEMINI_MODEL: "gemini-3.5-flash"
PROMPT_GUARD_BLOCK: "true"
PUSH_ENABLED: "false"
FCM_PROJECT_ID: "${GCP_PROJECT_ID}"
PDP_PROVIDER: "stub"
EOF
}

kore_write_frontend_env_file() {
  local path="$1"
  local api_url="${2:-$(api_run_url)}"
  local site_url="${3:-$PUBLIC_SITE_URL}"
  cat >"$path" <<EOF
NUXT_PUBLIC_API_BASE: "${api_url}"
NUXT_API_BASE: "${api_url}"
NUXT_PUBLIC_SITE_URL: "${site_url}"
EOF
}

kore_api_secrets() {
  local secrets
  secrets="DATABASE_URL=kore-database-url:latest"
  secrets+=",JWT_SIGNING_KEY=kore-jwt-signing-key:latest"
  secrets+=",TOTP_ENCRYPTION_KEY=kore-totp-encryption-key:latest"
  secrets+=",STRIPE_SECRET_KEY=kore-stripe-secret-key:latest"
  secrets+=",STRIPE_WEBHOOK_SECRET=kore-stripe-webhook-secret:latest"
  secrets+=",STRIPE_PUBLISHABLE_KEY=kore-stripe-publishable-key:latest"
  if kore_has_secret_version "$GEMINI_API_KEY_SECRET"; then
    secrets+=",GEMINI_API_KEY=${GEMINI_API_KEY_SECRET}:latest"
  fi
  if kore_has_secret_version "kore-oidc-google-client-secret"; then
    secrets+=",OIDC_GOOGLE_CLIENT_SECRET=kore-oidc-google-client-secret:latest"
  fi
  if kore_has_secret_version "kore-oidc-google-client-id"; then
    secrets+=",OIDC_GOOGLE_CLIENT_ID=kore-oidc-google-client-id:latest"
  fi
  if kore_has_secret_version "kore-pdp-api-key"; then
    secrets+=",PDP_API_KEY=kore-pdp-api-key:latest"
  fi
  if kore_has_secret_version "kore-pdp-webhook-secret"; then
    secrets+=",PDP_WEBHOOK_SECRET=kore-pdp-webhook-secret:latest"
  fi
  if kore_has_secret_version "kore-pennylane-api-token"; then
    secrets+=",PENNYLANE_API_TOKEN=kore-pennylane-api-token:latest"
  fi
  if kore_has_secret_version "kore-fcm-service-account"; then
    secrets+=",GOOGLE_APPLICATION_CREDENTIALS=kore-fcm-service-account:latest"
  fi
  printf '%s' "$secrets"
}

kore_migrate_secrets() {
  local secrets
  if gcloud secrets versions access latest \
    --secret=kore-migrate-database-url --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="DATABASE_URL=kore-migrate-database-url:latest"
  else
    echo "→ Job migrate : secret kore-migrate-database-url absent — fallback runtime" >&2
    secrets="DATABASE_URL=kore-database-url:latest"
  fi
  secrets+=",TOTP_ENCRYPTION_KEY=kore-totp-encryption-key:latest"
  printf '%s' "$secrets"
}

kore_seed_secrets() {
  printf '%s' "$(kore_api_secrets)"
}

kore_has_secret_version() {
  local secret="$1"
  gcloud secrets versions access latest \
    --secret="$secret" --project="$GCP_PROJECT_ID" >/dev/null 2>&1
}
