# Kafka + Debezium → ClickHouse Architecture

> Branch: `feature/kafka-debezium-architecture`  
> Target: server terpisah **minimal 4GB RAM, ideal 8GB+**. Jangan deploy stack ini ke VPS DDAG 908MB.

## Why this architecture

Reference `ProgrammerZamanNow/oltp-olap-demo` uses the classic CDC pipeline:

```text
OLTP DB ──CDC──▶ Debezium Connect ──events──▶ Kafka ──Kafka Engine / Sink──▶ ClickHouse
```

For our new project, this is the best option when the goal is:

- multi-source database support
- real CDC, not polling
- PostgreSQL WAL streaming
- MySQL binlog streaming
- SQL Server CDC support
- future Oracle/MongoDB expansion
- replayable event log
- scalable ingestion into ClickHouse

## Recommended topology

```text
                    ┌────────────────────┐
                    │ Source Databases   │
                    │ PG / MySQL / MSSQL │
                    └─────────┬──────────┘
                              │ CDC
                              ▼
┌─────────────┐       ┌──────────────────┐       ┌─────────────┐
│ Schema      │◀─────▶│ Kafka Connect    │──────▶│ Kafka       │
│ Registry    │       │ Debezium plugins │       │ Topics      │
└─────────────┘       └──────────────────┘       └──────┬──────┘
                                                         │
                                                         ▼
                                                 ┌──────────────┐
                                                 │ ClickHouse   │
                                                 │ Kafka Engine │
                                                 │ + MVs        │
                                                 └──────────────┘
```

## Service sizing

| Component | Minimum | Comfortable | Notes |
|---|---:|---:|---|
| Kafka KRaft | 1.5GB | 2-4GB | JVM, broker storage cache |
| Kafka Connect + Debezium | 1GB | 2GB | JVM, connector tasks |
| ClickHouse | 1GB | 2-4GB | depends on ingestion volume |
| Schema Registry | 512MB | 1GB | optional but recommended |
| Kafka UI | 256MB | 512MB | optional |

**Minimum server:** 4GB RAM.  
**Recommended:** 8GB RAM, 2-4 vCPU, SSD.

## Branch policy

- `archive/v1-cms-generator` — old native Go CMS + generator snapshot.
- `main` — stable repo state.
- `feature/kafka-debezium-architecture` — Kafka + Debezium heavy architecture work.

## Directory layout

```text
deploy/kafka-debezium/
├── README.md
├── docker-compose.yml
├── .env.example
├── connectors/
│   ├── postgres-source.json
│   ├── mysql-source.json
│   └── sqlserver-source.json
├── clickhouse/
│   ├── 001_database.sql
│   ├── 010_kafka_tables.sql
│   └── 020_materialized_views.sql
└── scripts/
    ├── register-connectors.sh
    ├── status.sh
    └── reset-demo.sh
```

## Quick start

```bash
cd deploy/kafka-debezium
cp .env.example .env
# edit .env for source DB credentials
podman compose up -d
./scripts/status.sh
./scripts/register-connectors.sh
```

## Important notes

1. Enable CDC on source DBs first.
2. Use separate low-privileged CDC users.
3. Do not expose Kafka/Connect/ClickHouse publicly without auth + network ACLs.
4. ClickHouse ingestion is done via Kafka Engine + Materialized Views.
5. For SQL Server, CDC must be enabled per database and per table.
