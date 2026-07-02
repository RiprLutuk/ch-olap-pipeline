#!/usr/bin/env bash
set -euo pipefail

COMPOSE="${COMPOSE:-podman compose}"

$COMPOSE ps

echo "--- connect ---"
curl -fsS http://127.0.0.1:${HOST_CONNECT_PORT:-8083}/connectors || true

echo "--- clickhouse ---"
curl -fsS http://127.0.0.1:${HOST_CLICKHOUSE_HTTP_PORT:-8123}/ping || true
