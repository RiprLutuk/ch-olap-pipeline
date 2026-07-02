# PeerDB setup notes for ch-olap-pipeline

This project can use an existing external ClickHouse and external source databases.

## Goal
Stream data from PostgreSQL / MySQL / SQL Server / other SQL sources into ClickHouse **without** running Kafka/Debezium on a 908MB VPS.

## Recommended pattern
- Run **CMS** and optional **generator** on the VPS as native systemd services.
- Use **existing external ClickHouse** if you already have one.
- Use **source DBs** (RDS / on-prem / other VM) as upstreams.
- Use a dedicated CDC/replication tool outside the tiny VPS when possible.

## Important note about PeerDB
PeerDB is strongest for **Postgres -> ClickHouse** replication.
For MySQL / SQL Server you may need a different CDC tool or a separate ingestion path.

## Suggested options by source
| Source | Best path to ClickHouse |
|---|---|
| PostgreSQL | PeerDB / ClickPipes / custom sync job |
| MySQL | custom sync job, Airbyte, Debezium, or managed pipeline |
| SQL Server | custom sync job, Debezium, Airbyte, or managed pipeline |

## Minimal VPS-safe approach
If RAM is tight, do **not** run Kafka + Connect + ClickHouse + source DBs together on the same VPS.
Instead:
1. Keep source DBs external.
2. Keep ClickHouse external or native-local only.
3. Run lightweight sync workers on schedule.
4. Expose only CMS via Caddy.

## Example environment for external ClickHouse
```bash
export CMS_CLICKHOUSE_URL=http://YOUR-CH-HOST:8123
export CMS_CLICKHOUSE_USER=analytics
export CMS_CLICKHOUSE_PASSWORD=YOUR_PASSWORD
sudo -E ./deploy/vps/install-native.sh --external-clickhouse
```

## Example environment for local ClickHouse + external Postgres/MySQL
```bash
export CMS_PG_DSN='postgres://user:pass@your-rds:5432/dbname?sslmode=require'
export CMS_MYSQL_DSN='user:pass@tcp(your-rds:3306)/dbname'
export GEN_PG_DSN="$CMS_PG_DSN"
export GEN_MYSQL_DSN="$CMS_MYSQL_DSN"
sudo -E ./deploy/vps/install-native.sh
```

## Suggested future improvement
Build dedicated sync commands:
- `cmd/sync-pg-to-ch`
- `cmd/sync-mysql-to-ch`
- `cmd/sync-sqlserver-to-ch`

That would fit the user's no-container / low-RAM requirement better than Kafka stack.
