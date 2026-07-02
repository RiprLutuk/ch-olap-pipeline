#!/usr/bin/env bash
# install.sh - install ClickHouse + CMS + Generator as native systemd services on Ubuntu/Debian
# Usage: sudo ./install.sh [--no-clickhouse] [--no-cms] [--no-generator]
set -euo pipefail

log()  { printf "\033[1;36m[install]\033[0m %s\n" "$*"; }
warn() { printf "\033[1;33m[warn]\033[0m %s\n" "$*" >&2; }
fail() { printf "\033[1;31m[fail]\033[0m %s\n" "$*" >&2; exit 1; }

[[ $EUID -eq 0 ]] || fail "must run as root (use sudo)"

WITH_CH=1; WITH_CMS=1; WITH_GEN=1
for a in "$@"; do
  case "$a" in
    --no-clickhouse) WITH_CH=0 ;;
    --no-cms)        WITH_CMS=0 ;;
    --no-generator)  WITH_GEN=0 ;;
    -h|--help)
      sed -n "2,6p" "$0"; exit 0 ;;
  esac
done

# Detect architecture
arch=$(dpkg --print-architecture)
[[ $arch == amd64 || $arch == arm64 ]] || fail "unsupported arch: $arch"

# Detect dist
. /etc/os-release
[[ $ID == ubuntu || $ID == debian ]] || fail "needs Ubuntu or Debian"

# Common
log "apt update + base packages"
apt-get update -y
apt-get install -y curl gnupg ca-certificates apt-transport-https

# Create service user
id -u chapp >/dev/null 2>&1 || useradd --system --home /var/lib/chapp --shell /usr/sbin/nologin chapp
install -d -m 0755 -o chapp -g chapp /var/lib/chapp

# === ClickHouse ===
if [[ $WITH_CH -eq 1 ]]; then
  # Generate random password
CH_PASS=$(openssl rand -base64 18 | tr -d "=+/")
CH_HASH=$(echo -n "$CH_PASS" | sha256sum | awk '{print $1}')
log "ClickHouse analytics user password (saved to /etc/ch-olap/cms.env) -> $CH_PASS"

log "install ClickHouse"
  . /etc/os-release
  mkdir -p /etc/apt/keyrings
  curl -fsSL https://packages.clickhouse.com/deb/lts/repo.gpg | gpg --dearmor -o /etc/apt/keyrings/clickhouse-keyring.gpg
  echo "deb [signed-by=/etc/apt/keyrings/clickhouse-keyring.gpg] https://packages.clickhouse.com/deb/lts $VERSION_CODENAME main" > /etc/apt/sources.list.d/clickhouse.list
  apt-get update -y
  DEBIAN_FRONTEND=noninteractive apt-get install -y clickhouse-server clickhouse-client
  log "configure ClickHouse"
  install -d -m 0750 -o clickhouse -g clickhouse /var/lib/clickhouse /var/log/clickhouse-server
  cat > /etc/clickhouse-server/config.d/pipeline.xml <<EOF
<clickhouse>
  <listen_host>127.0.0.1</listen_host>
  <max_connections>200</max_connections>
  <max_concurrent_queries>50</max_concurrent_queries>
  <uncompressed_cache_size>268435456</uncompressed_cache_size>
  <mark_cache_size>536870912</mark_cache_size>
  <path>/var/lib/clickhouse/</path>
  <tmp_path>/var/lib/clickhouse/tmp/</tmp_path>
  <user_files_path>/var/lib/clickhouse/user_files/</user_files_path>
  <users_config>users.xml</users_config>
  <default_profile>default</default_profile>
  <default_database>default</default_database>
  <timezone>UTC</timezone>
  <mlock_executable>false</mlock_executable>
  <max_server_memory_usage>0</max_server_memory_usage>
  <max_thread_pool_size>10000</max_thread_pool_size>
  <total_memory_profiler_step>4194304</total_memory_profiler_step>
  <merges_mutations_memory_usage_soft_limit>0</merges_mutations_memory_usage_soft_limit>
  <profiles>
    <default/>
  </profiles>
  <users>
    <analytics>
      <password remove="1"/><password_sha256_hex>$CH_HASH</password_sha256_hex>
      <networks incl="networks" replace="replace">
        <ip>::/0</ip>
      </networks>
      <profile>default</profile>
      <quota>default</quota>
      <access_management>1</access_management>
    </analytics>
  </users>
  <quotas><default><interval><duration>3600</duration><queries>0</queries><errors>0</errors><result_rows>0</result_rows><read_rows>0</read_rows><execution_time>0</execution_time></interval></default></quotas>
</clickhouse>
EOF
  systemctl enable --now clickhouse-server
  sleep 5
  clickhouse-client --query "SELECT version()" || warn "ClickHouse not responding yet; check journalctl -u clickhouse-server"
fi

# === CMS ===
if [[ $WITH_CMS -eq 1 ]]; then
  log "build CMS"
  apt-get install -y golang-go
  install -d /opt/ch-olap
  rsync -a --delete "$(dirname "$0")/../../" /opt/ch-olap/src/ || cp -r "$(dirname "$0")/../../." /opt/ch-olap/src/
  cd /opt/ch-olap/src
  go build -trimpath -ldflags="-s -w" -o /usr/local/bin/ch-cms ./cmd/cms
  log "write CMS systemd"
  cat > /etc/systemd/system/ch-cms.service <<EOF
[Unit]
Description=CH OLAP Pipeline CMS
After=network-online.target
Wants=network-online.target

[Service]
User=chapp
Group=chapp
WorkingDirectory=/var/lib/chapp
EnvironmentFile=-/etc/ch-olap/cms.env
ExecStart=/usr/local/bin/ch-cms
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/var/lib/chapp
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
  install -d /etc/ch-olap
  cat > /etc/ch-olap/cms.env <<EOF
CMS_HTTP_ADDR=:8080
CMS_GENERATOR_URL=http://127.0.0.1:8081
CMS_CLICKHOUSE_URL=http://127.0.0.1:8123
CMS_CLICKHOUSE_USER=analytics
CMS_CLICKHOUSE_PASSWORD=analytics
CMS_PG_DSN=postgres://shop:***@database-1.cwt8i4qk6wnt.us-east-1.rds.amazonaws.com:5432/shop?sslmode=require
CMS_MYSQL_DSN=shop:***@tcp(database-2.cwt8i4qk6wnt.us-east-1.rds.amazonaws.com:3306)/shop
EOF
  systemctl daemon-reload
  systemctl enable --now ch-cms
fi

# === Generator ===
if [[ $WITH_GEN -eq 1 ]]; then
  log "build Generator"
  cd /opt/ch-olap/src
  go build -trimpath -ldflags="-s -w" -o /usr/local/bin/ch-gen ./cmd/generator
  cat > /etc/systemd/system/ch-gen.service <<EOF
[Unit]
Description=CH OLAP Pipeline Generator
After=network-online.target
Wants=network-online.target

[Service]
User=chapp
Group=chapp
WorkingDirectory=/var/lib/chapp
EnvironmentFile=-/etc/ch-olap/gen.env
ExecStart=/usr/local/bin/ch-gen
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
  cat > /etc/ch-olap/gen.env <<EOF
GEN_HTTP_ADDR=:8081
GEN_TICK_MS=1000
GEN_WORKERS=2
GEN_PG_DSN=postgres://shop:***@database-1.cwt8i4qk6wnt.us-east-1.rds.amazonaws.com:5432/shop?sslmode=require
GEN_MYSQL_DSN=shop:***@tcp(database-2.cwt8i4qk6wnt.us-east-1.rds.amazonaws.com:3306)/shop
EOF
  systemctl daemon-reload
  systemctl enable --now ch-gen
fi

log "done"
log "check status: systemctl status ch-cms ch-gen clickhouse-server"
