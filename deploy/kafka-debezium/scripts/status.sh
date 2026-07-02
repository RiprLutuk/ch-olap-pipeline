#!/usr/bin/env bash
set -euo pipefail

COMPOSE="${COMPOSE:-podman compose}"
CONNECT_URL="${CONNECT_URL:-http://127.0.0.1:${HOST_CONNECT_PORT:-8083}}"
CLICKHOUSE_URL="${CLICKHOUSE_URL:-http://127.0.0.1:${HOST_CLICKHOUSE_HTTP_PORT:-8123}}"

if command -v sh >/dev/null 2>&1; then
  echo "==> compose ps"
  $COMPOSE ps || true
fi

echo
printf '==> connect health: '
if curl -fsS "$CONNECT_URL/connectors" >/dev/null; then
  echo "ok"
  curl -fsS "$CONNECT_URL/connectors"
else
  echo "unreachable"
fi

echo
printf '==> clickhouse health: '
if curl -fsS "$CLICKHOUSE_URL/ping" >/dev/null; then
  echo "ok"
  curl -fsS "$CLICKHOUSE_URL/ping"
else
  echo "unreachable"
fi

echo
printf '==> kafka-ui health: '
if curl -fsS "http://127.0.0.1:${HOST_KAFKA_UI_PORT:-8088}" >/dev/null; then
  echo "ok"
else
  echo "unreachable"
fi
