#!/usr/bin/env sh
set -euo pipefail

: "${URL:=http://localhost:8080/metrics}"
: "${METRIC:=http_requests_total}"
: "${INTERVAL:=1}"

echo "Watching '$METRIC' from $URL every ${INTERVAL}s..."
while :; do
  ts=$(date +%H:%M:%S)
  # Match either `metric{...}` or `metric ` (no labels)
  curl -sf "$URL" \
  | grep -E "^cilium_operator_lbipam" \
  | sed "s/^/[$ts] /" || echo "[$ts] (no match or scrape error)"
  sleep "$INTERVAL"
done
