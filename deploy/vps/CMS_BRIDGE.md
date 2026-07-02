# CMS bridge: external vs self-hosted ClickHouse

The CMS already supports both modes through environment variables.

## Self-hosted ClickHouse
Use when ClickHouse runs on the same VPS.

```env
CMS_CLICKHOUSE_URL=http://127.0.0.1:8123
CMS_CLICKHOUSE_USER=analytics
CMS_CLICKHOUSE_PASSWORD=analytics
```

## External ClickHouse
Use when you already have a ClickHouse server elsewhere.

```env
CMS_CLICKHOUSE_URL=http://YOUR-CLICKHOUSE-HOST:8123
CMS_CLICKHOUSE_USER=analytics
CMS_CLICKHOUSE_PASSWORD=YOUR_SECRET
```

## Behavior
The CMS does not need code changes to detect the mode.
It simply uses the configured HTTP endpoint:
- `127.0.0.1` => self-hosted/local mode
- remote host => external mode

## Install examples
### Local CH
```bash
sudo ./deploy/vps/install-native.sh
```

### External CH
```bash
export CMS_CLICKHOUSE_URL=http://YOUR-CLICKHOUSE-HOST:8123
export CMS_CLICKHOUSE_USER=analytics
export CMS_CLICKHOUSE_PASSWORD=YOUR_SECRET
sudo -E ./deploy/vps/install-native.sh --external-clickhouse
```

## Optional Caddy block
```caddy
cms.example.com {
    encode gzip zstd
    reverse_proxy 127.0.0.1:8084
}
```

## Security note
Prefer keeping ClickHouse bound privately and only exposing CMS through Caddy.
