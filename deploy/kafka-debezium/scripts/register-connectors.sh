#!/usr/bin/env bash
set -euo pipefail

CONNECT_URL="${CONNECT_URL:-http://127.0.0.1:${HOST_CONNECT_PORT:-8083}}"

for f in connectors/*.json; do
  echo "registering $f"
  envsubst < "$f" | curl -fsS -X POST "$CONNECT_URL/connectors" \
    -H 'Content-Type: application/json' \
    --data @-
  echo
done
