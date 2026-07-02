# ch-olap-pipeline

A high-performance, multi-connection **OLTP → OLAP** data pipeline using **Go**, **Debezium**, **Apache Kafka**, and **ClickHouse**.

Inspired by [`ProgrammerZamanNow/oltp-olap-demo`](https://github.com/ProgrammerZamanNow/oltp-olap-demo), but rebuilt from scratch with:

| Concern | Reference (Java/Spring) | This project (Go) |
|---|---|---|
| Memory footprint | ~300–500 MB per service | ~10–25 MB per service |
| Concurrency model | Thread pools | Native goroutines |
| Database support | PostgreSQL only | PostgreSQL **+ MySQL** (multi-DB) |
| Generator speed | JPA/Hibernate (slow) | Bulk inserts via `pgx` / `go-sql-driver/mysql` |
| Config | Hardcoded in YAML | 12-factor, env-driven |
| Topology | Monolithic compose | Modular compose + profiles |

---

## Architecture

```text
┌─────────────────────────────────────────────────────────────────────┐
│  Go Generator (cmd/generator)                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │ Worker Pool  │  │ Worker Pool  │  │ Worker Pool  │  (goroutines) │
│  │  → Postgres  │  │  → MySQL     │  │  → (future)  │               │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘               │
└─────────┼──────────────────┼──────────────────┼─────────────────────┘
          │                  │                  │
          ▼                  ▼                  ▼
   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
   │ PostgreSQL  │    │   MySQL     │    │   (other)   │   ← OLTP sources
   │ (WAL)       │    │  (Binlog)   │    │             │
   └──────┬──────┘    └──────┬──────┘    └─────────────┘
          │ logical repl     │ binlog
          ▼                  ▼
   ┌──────────────────────────────────────────────┐
   │  Debezium (Kafka Connect) — single cluster   │
   │  • postgres-connector                         │
   │  • mysql-connector                            │
   └──────────────────┬───────────────────────────┘
                      │  topics: shop.public.*, shop_mysql.*
                      ▼
   ┌──────────────────────────────────────────────┐
   │  Apache Kafka (KRaft, single node)           │
   └──────────────────┬───────────────────────────┘
                      │
                      ▼
   ┌──────────────────────────────────────────────┐
   │  ClickHouse                                  │
   │   • Kafka Engine tables (per topic)          │
   │   • Materialized Views → ReplacingMergeTree  │
   │   • AggregatingMergeTree for rollups         │
   └──────────────────────────────────────────────┘
```

---

## Quick start

```bash
# 1. Bring up infra (Postgres, MySQL, Kafka, Connect, ClickHouse, Generator)
docker compose up -d

# 2. Register Debezium connectors
make register

# 3. Verify CDC is RUNNING
make status

# 4. Drive the generator (POST to API)
curl -X POST http://localhost:8080/api/v1/generator/start \
  -H 'Content-Type: application/json' \
  -d '{"workers": 8, "tick_ms": 1000, "targets": ["postgres", "mysql"]}'

# 5. Query ClickHouse
docker exec -it olap-clickhouse clickhouse-client \
  --user analytics --password analytics --database shop_analytics \
  --query "SELECT status, count() FROM orders GROUP BY status"
```

---

## Project layout

```
ch-olap-pipeline/
├── cmd/
│   └── generator/         # Go binary entry point
├── internal/
│   ├── api/               # HTTP handlers
│   ├── db/                # Connection pool manager
│   ├── generator/         # Worker pool + e-commerce simulator
│   └── model/             # Shared structs
├── deploy/
│   ├── postgres/init.sql
│   ├── mysql/init.sql
│   ├── clickhouse/init.sql
│   └── debezium/          # Connector JSONs
├── config/                # App config (env-driven)
├── scripts/               # Ops scripts
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## License

MIT
