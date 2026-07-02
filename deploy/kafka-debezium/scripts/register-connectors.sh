#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONNECT_DIR="$ROOT_DIR/connectors"
CONNECT_URL="${CONNECT_URL:-http://127.0.0.1:${HOST_CONNECT_PORT:-8083}}"
DRY_RUN="${DRY_RUN:-0}"

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

if ! command -v envsubst >/dev/null 2>&1; then
  echo "error: envsubst is required (gettext-base / gettext)" >&2
  exit 1
fi

shopt -s nullglob
files=("$CONNECT_DIR"/*.json)

if [ ${#files[@]} -eq 0 ]; then
  echo "error: no connector json files found in $CONNECT_DIR" >&2
  exit 1
fi

for f in "${files[@]}"; do
  echo "==> rendering $(basename "$f")"
  rendered="$(envsubst < "$f")"

  if printf '%s' "$rendered" | grep -q '\${'; then
    echo "error: unresolved environment variables remain in $(basename "$f")" >&2
    exit 1
  fi

  printf '%s' "$rendered" | python3 -m json.tool >/dev/null

  if [ "$DRY_RUN" = "1" ]; then
    echo "$rendered" | python3 -m json.tool
    echo "-- dry-run only, not posting to Kafka Connect --"
    continue
  fi

  echo "==> registering $(basename "$f") to $CONNECT_URL"
  printf '%s' "$rendered" | curl -fsS -X POST "$CONNECT_URL/connectors" \
    -H 'Content-Type: application/json' \
    --data @-
  echo
  echo
 done
