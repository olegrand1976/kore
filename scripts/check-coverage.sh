#!/usr/bin/env bash
# Gate de couverture sur la couche domain (règles métier pures).
# Les couches app/adapters restent couvertes par les tests d'intégration et smoke.
# Échoue si la couverture agrégée passe sous COVERAGE_THRESHOLD (défaut 40%).
set -euo pipefail

THRESHOLD="${COVERAGE_THRESHOLD:-40}"
PROFILE="${COVERAGE_PROFILE:-coverage.out}"

# Packages domain MVP (hors IA et modules squelettes post-MVP).
SKELETON_MODULES='admin|ett|integrations|invoicing|maintenance|reporting|ssii|support'
PKGS=$(go list ./... | grep -E '/modules/[^/]+/domain$' | grep -v '/modules/ai/' | grep -vE "/modules/(${SKELETON_MODULES})/" | paste -sd, -)
if [ -z "$PKGS" ]; then
  echo "no domain packages found" >&2
  exit 1
fi

go test -covermode=atomic -coverpkg="$PKGS" -coverprofile="$PROFILE" $(echo "$PKGS" | tr ',' ' ')

TOTAL=$(go tool cover -func="$PROFILE" | awk '/^total:/ {gsub("%","",$3); print $3}')
echo "Couverture domain agrégée : ${TOTAL}% (seuil ${THRESHOLD}%)"

awk -v total="$TOTAL" -v threshold="$THRESHOLD" 'BEGIN {
  if (total + 0 < threshold + 0) {
    printf("ÉCHEC : couverture %.1f%% < seuil %.1f%%\n", total, threshold) > "/dev/stderr"
    exit 1
  }
  printf("OK : couverture %.1f%% >= seuil %.1f%%\n", total, threshold)
}'
