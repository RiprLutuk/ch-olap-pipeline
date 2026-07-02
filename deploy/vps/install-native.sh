#!/usr/bin/env bash
# Native VPS installer for ch-olap-pipeline (no containers).
# Installs optional local ClickHouse plus CMS/Generator as systemd services.
set -euo pipefail

log(){ printf '\033[1;36m[native-install]\033[0m %s\n' "$*"; }
fail(){ printf '\033[1;31m[fail]\033[0m %s\n' "$*" >&2; exit 1; }

[[ ${EUID:-$(id -u)} -eq 0 ]] || fail "run as root: sudo $0"

INSTALL_CLICKHOUSE=1
INSTALL_CMS=1
INSTALL_GENERATOR=1
for arg in "$@"; do
  case "$arg" in
    --external-clickhouse|--no-clickhouse) INSTALL_CLICKHOUSE=0 ;;
    --no-cms) INSTALL_CMS=0 ;;
    --no-generator) INSTALL_GENERATOR=0 ;;
    -h|--help)
      cat <<HELP
Usage: sudo ./deploy/vps/install-native.sh [--external-clickhouse] [--no-cms] [--no-generator]

Environment overrides before running:
  CMS_CLICKHOUSE_URL       default http://127.0.0.1:8123
  CMS_CLICKHOUSE_USER      default analytics
  CMS_CLICKHOUSE_PASSWORD  default analytics
  CMS_GENERATOR_URL        default http://127.0.0.1:8085
  CMS_PG_DSN               optional external Postgres DSN
  CMS_MYSQL_DSN            optional external MySQL DSN
  GEN_HTTP_ADDR            default :8085
  GEN_PG_DSN               optional external Postgres DSN
  GEN_MYSQL_DSN            optional external MySQL DSN
  GEN_TICK_MS              default 1000
  GEN_WORKERS              default 2
HELP
      exit 0 ;;
    *) fail "unknown arg: $arg" ;;
  esac
done

command -v apt-get >/dev/null || fail "Ubuntu/Debian required"
apt-get update -y
apt-get install -y ca-certificates curl gnupg git rsync golang-go
install -d /opt/ch-olap /etc/ch-olap /var/lib/chapp
id -u chapp >/dev/null 2>&1 || useradd --system --home /var/lib/chapp --shell /usr/sbin/nologin chapp
chown chapp:chapp /var/lib/chapp

if [[ $INSTALL_CLICKHOUSE -eq 1 ]]; then
  log "install ClickHouse native package"
  install -d -m 0755 /etc/apt/keyrings
  curl -fsSL https://packages.clickhouse.com/deb/lts/Release.key | gpg --dearmor -o /etc/apt/keyrings/clickhouse-keyring.gpg
  echo "deb [signed-by=/etc/apt/keyrings/clickhouse-keyring.gpg] https://packages.clickhouse.com/deb stable main" > /etc/apt/sources.list.d/clickhouse.list
  apt-get update -y
  DEBIAN_FRONTEND=noninteractive apt-get install -y clickhouse-server clickhouse-client
  install -d -m 0755 /etc/clickhouse-server/config.d /etc/clickhouse-server/users.d
  cat >/etc/clickhouse-server/config.d/listen-local.xml <<'XML'
<clickhouse>
  <listen_host>127.0.0.1</listen_host>
</clickhouse>
XML
  cat >/etc/clickhouse-server/users.d/analytics.xml <<'XML'
<clickhouse>
  <users>
    <analytics>
      <password>analytics</password>
      <networks>
        <ip>127.0.0.1</ip>
        <ip>::1</ip>
      </networks>
      <profile>default</profile>
      <quota>default</quota>
      <access_management>1</access_management>
    </analytics>
  </users>
</clickhouse>
XML
  systemctl enable --now clickhouse-server
  sleep 3
  clickhouse-client --query "CREATE DATABASE IF NOT EXISTS shop_analytics"
fi

log "copy source and build binaries"
rsync -a --delete --exclude .git ./ /opt/ch-olap/src/
cd /opt/ch-olap/src
if [[ $INSTALL_CMS -eq 1 ]]; then
  go build -trimpath -ldflags='-s -w' -o /usr/local/bin/ch-cms ./cmd/cms
fi
if [[ $INSTALL_GENERATOR -eq 1 ]]; then
  go build -trimpath -ldflags='-s -w' -o /usr/local/bin/ch-gen ./cmd/generator
fi

cat >/etc/ch-olap/cms.env <<EOF
CMS_HTTP_ADDR=:8084
CMS_GENERATOR_URL=${CMS_GENERATOR_URL:-http://127.0.0.1:8085}
CMS_CLICKHOUSE_URL=${CMS_CLICKHOUSE_URL:-http://127.0.0.1:8123}
CMS_CLICKHOUSE_USER=${CMS_CLICKHOUSE_USER:-analytics}
CMS_CLICKHOUSE_PASSWORD=${CMS_CLICKHOUSE_PASSWORD:-analytics}
CMS_PG_DSN=${CMS_PG_DSN:-postgres://shop:shop@127.0.0.1:5432/shop?sslmode=disable}
CMS_MYSQL_DSN=${CMS_MYSQL_DSN:-shop:shop@tcp(127.0.0.1:3306)/shop}
EOF
chmod 0600 /etc/ch-olap/cms.env

cat >/etc/ch-olap/gen.env <<EOF
GEN_HTTP_ADDR=${GEN_HTTP_ADDR:-:8085}
GEN_TICK_MS=${GEN_TICK_MS:-1000}
GEN_WORKERS=${GEN_WORKERS:-2}
GEN_PG_DSN=${GEN_PG_DSN:-postgres://shop:shop@127.0.0.1:5432/shop?sslmode=disable}
GEN_MYSQL_DSN=${GEN_MYSQL_DSN:-shop:shop@tcp(127.0.0.1:3306)/shop}
EOF
chmod 0600 /etc/ch-olap/gen.env

if [[ $INSTALL_CMS -eq 1 ]]; then
  cat >/etc/systemd/system/ch-cms.service <<'UNIT'
[Unit]
Description=CH OLAP Pipeline CMS
After=network-online.target
Wants=network-online.target

[Service]
User=chapp
Group=chapp
EnvironmentFile=/etc/ch-olap/cms.env
ExecStart=/usr/local/bin/ch-cms
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
UNIT
fi

if [[ $INSTALL_GENERATOR -eq 1 ]]; then
  cat >/etc/systemd/system/ch-gen.service <<'UNIT'
[Unit]
Description=CH OLAP Pipeline Generator
After=network-online.target
Wants=network-online.target

[Service]
User=chapp
Group=chapp
EnvironmentFile=/etc/ch-olap/gen.env
ExecStart=/usr/local/bin/ch-gen
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
UNIT
fi

systemctl daemon-reload
[[ $INSTALL_CMS -eq 1 ]] && systemctl enable --now ch-cms
[[ $INSTALL_GENERATOR -eq 1 ]] && systemctl enable --now ch-gen

log "done"
log "CMS: http://127.0.0.1:8084"
log "Status: systemctl status ch-cms ch-gen clickhouse-server"
