#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONNECT_DIR="$ROOT_DIR/connectors"
CONNECT_URL="${CONNECT_URL:-http://127.0.0.1:${HOST_CONNECT_PORT:-8083}}"
DRY_RUN="${DRY_RUN:-0}"

required_env=(
  TOPIC_PREFIX
  PG_HOST PG_PORT PG_USER PG_PASSWORD PG_DB PG_TABLE_INCLUDE
  MYSQL_HOST MYSQL_PORT MYSQL_USER MYSQL_PASSWORD MYSQL_DB MYSQL_TABLE_INCLUDE
  MSSQL_HOST MSSQL_PORT MSSQL_USER MSSQL_PASSWORD MSSQL_DB MSSQL_TABLE_INCLUDE
)

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

if ! command -v envsubst >/dev/null 2>&1; then
  echo "error: envsubst is required (gettext-base / gettext)" >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "error: python3 is required for json validation" >&2
  exit 1
fi

missing=()
for name in "${required_env[@]}"; do
  if [ -z "${!name:-}" ]; then
    missing+=("$name")
  fi
done

if [ ${#missing[@]} -gt 0 ]; then
  echo "error: required environment variables are missing or empty:" >&2
  printf '  - %s\n' "${missing[@]}" >&2
  echo "hint: copy .env.example to .env, fill it, then run: set -a; source .env; set +a" >&2
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
