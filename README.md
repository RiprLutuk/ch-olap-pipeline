# ch-olap-pipeline

> A practical open-source **OLTP → OLAP** data pipeline built around **real Change Data Capture (CDC)**, designed for the many real-world cases where you are stuck with **heterogeneous RDBMS** sources and need a clean, **observable** path into **ClickHouse** for analytics.

[![CI](https://github.com/RiprLutuk/ch-olap-pipeline/actions/workflows/ci.yml/badge.svg)](https://github.com/RiprLutuk/ch-olap-pipeline/actions/workflows/ci.yml)
[![Docs](https://github.com/RiprLutuk/ch-olap-pipeline/actions/workflows/pages.yml/badge.svg)](https://github.com/RiprLutuk/ch-olap-pipeline/actions/workflows/pages.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Debezium](https://img.shields.io/badge/Debezium-2.7-blueviolet)](https://debezium.io)
[![Kafka](https://img.shields.io/badge/Apache_Kafka-KRaft-black?logo=apachekafka)](https://kafka.apache.org)
[![ClickHouse](https://img.shields.io/badge/ClickHouse-24.x-yellow?logo=clickhouse&logoColor=black)](https://clickhouse.com)

> Inspired by [`ProgrammerZamanNow/oltp-olap-demo`](https://github.com/ProgrammerZamanNow/oltp-olap-demo), but rebuilt for the real world: multiple database engines, modular adapters, production-first docs.

---

## Why this project exists

In real DBA / infra / platform work, the painful truth is:

- You rarely have **only one database** — there's PostgreSQL, MySQL, SQL Server, and at some point Oracle or Db2 too.
- **Polling-based ETL breaks** under scale, especially with **DELETE** semantics and high-frequency changes.
- Off-the-shelf CDC products are often **expensive, opaque, and vendor-locked**.
- The classic **Java + Kafka + Debezium + ClickHouse** stack has excellent primitives, but the **operational knowledge to run it well is scattered**.

This project tries to close that gap with a **transparent, modular, runnable reference** that:

- speaks to **many RDBMS sources**, not just one
- makes **CDC mechanics** explicit and observable
- ships with **install paths, gap recovery, troubleshooting, and observability**
- can run on a **small VM for learning** and scale out to a **multi-host production shape**

---

## Mission

> Build an open-source platform for engineers, DBAs, data platform teams, and backend/infra practitioners who need to replicate data changes from many OLTP RDBMS sources into analytical systems such as ClickHouse with **real CDC**, operational clarity, and production-grade troubleshooting.

Read the full vision: [`docs/open-source-vision.md`](docs/open-source-vision.md)

---

## Supported sources

| Source DB | CDC mechanism | Status |
|---|---|---|
| **PostgreSQL** | WAL logical replication | 🟢 demo-ready |
| **MySQL** | binlog row-based | 🟢 demo-ready |
| **SQL Server** | native CDC | 🟢 demo-ready |
| **MariaDB** | binlog compatible | 🟡 beta |
| **Oracle** | LogMiner / XStream | 🟡 beta |
| **IBM Db2** | transaction log capture | 📋 planned |
| **CockroachDB** | changefeeds | 📋 planned |
| **Vitess / PlanetScale** | MySQL-compatible path | 📋 planned |
| **YugabyteDB** | PG-wire compatible variant | 🔬 research |
| **SQLite** | batch bridge | 🔬 research |
| **Firebird** | trigger/log bridge | 🔬 research |
| **SAP HANA** | SDI path | 🔬 research |

See: [`docs/supported-rdbms-matrix.md`](docs/supported-rdbms-matrix.md)

---

## Supported sinks

| Sink | Status |
|---|---|
| **ClickHouse** | 🟢 active |
| BigQuery | 📋 planned |
| Snowflake | 📋 planned |
| Apache Iceberg | 📋 planned |
| Delta Lake | 📋 planned |
| PostgreSQL warehouse | 📋 planned |

---

## Architecture at a glance

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                          SOURCE LAYER (OLTP)                            │
│                                                                         │
│   ┌────────────┐  ┌────────────┐  ┌─────────────┐  ┌────────────┐      │
│   │ PostgreSQL │  │   MySQL    │  │ SQL Server  │  │   Oracle   │      │
│   │   WAL      │  │  binlog    │  │  CDC job    │  │  LogMiner  │      │
│   └─────┬──────┘  └─────┬──────┘  └──────┬──────┘  └─────┬──────┘      │
└─────────┼───────────────┼────────────────┼────────────────┼────────────┘
          │               │                │                │
          ▼               ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                  CAPTURE LAYER (Debezium Connect)                      │
│                                                                         │
│   Kafka Connect cluster with one source connector per database engine.  │
│   Modular: each adapter lives in its own config, easy to extend.        │
└──────────────────────────────────┬──────────────────────────────────────┘
                                   │ JSON events
                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    EVENT BUS LAYER (Apache Kafka)                        │
│                                                                         │
│   KRaft single-node (dev) → 3-node cluster (prod)                       │
│   Topics: oltpdemo.{db}.{schema}.{table}                                │
│   Schema Registry optional but recommended                              │
└──────────────────────────────────┬──────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    ANALYTICAL LAYER (ClickHouse)                         │
│                                                                         │
│   • Kafka Engine tables (one per topic)                                 │
│   • Materialized views → ReplacingMergeTree (deduplication safe)        │
│   • AggregatingMergeTree for rollups                                    │
│   • Distributed tables for multi-shard scale                            │
└─────────────────────────────────────────────────────────────────────────┘
```

For the deep-dive: [`deploy/kafka-debezium/README.md`](deploy/kafka-debezium/README.md) — covers CDC mechanics, multi-server topology, install paths, and troubleshooting.

---

## Quick start

You need a host with **at least 4GB RAM (8GB recommended)**. This project is **not** designed for very small VPS like 1GB or less.

```bash
# 1. Clone
git clone https://github.com/RiprLutuk/ch-olap-pipeline.git
cd ch-olap-pipeline

# 2. Choose architecture branch
git checkout feature/kafka-debezium-architecture

# 3. Configure
cd deploy/kafka-debezium
cp .env.example .env
# edit .env with your source DB credentials, brokers, etc.

# 4. Bring up the stack
podman compose up -d     # or: docker compose up -d

# 5. Wait ~30s, then verify
./scripts/status.sh

# 6. Register Debezium source connectors
./scripts/register-connectors.sh
```

Open **Kafka UI** at `http://your-host:8088` to inspect topics, schemas, and connector health.

---

## Repository layout

```text
ch-olap-pipeline/
├── README.md                              ← you are here
├── LICENSE                                ← MIT
├── CONTRIBUTING.md                        ← how to contribute
├── CODE_OF_CONDUCT.md                     ← community standards
├── docs/
│   ├── open-source-vision.md              ← mission & positioning
│   ├── supported-rdbms-matrix.md          ← source DB coverage roadmap
│   └── architecture-decisions.md          ← ADRs (why we chose what)
├── deploy/
│   ├── kafka-debezium/                    ← active architecture (heavy)
│   │   ├── README.md                      ← deep-dive guide
│   │   ├── docker-compose.yml
│   │   ├── .env.example
│   │   ├── connectors/
│   │   │   ├── postgres-source.json
│   │   │   ├── mysql-source.json
│   │   │   └── sqlserver-source.json
│   │   ├── clickhouse/
│   │   │   ├── 001_database.sql
│   │   │   ├── 010_kafka_tables.sql
│   │   │   └── 020_materialized_views.sql
│   │   └── scripts/
│   │       ├── register-connectors.sh
│   │       └── status.sh
│   └── vps/                               ← native installer (legacy, archived)
└── cmd/                                   ← Go services (legacy, archived)
```

Branches:
- `main` — stable, documentation-first
- `feature/kafka-debezium-architecture` — active development of the heavy CDC stack
- `archive/v1-cms-generator` — legacy native Go CMS + generator snapshot (read-only)

---

## Documentation map

| Document | Purpose |
|---|---|
| [`README.md`](README.md) | Project landing page, quickstart, links |
| [`docs/open-source-vision.md`](docs/open-source-vision.md) | Mission, principles, positioning |
| [`docs/supported-rdbms-matrix.md`](docs/supported-rdbms-matrix.md) | Source DB & sink coverage roadmap |
| [`docs/architecture-decisions.md`](docs/architecture-decisions.md) | Why Kafka+Debezium, why ClickHouse, why modular |
| [`deploy/kafka-debezium/README.md`](deploy/kafka-debezium/README.md) | Full architecture guide, install, troubleshooting |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | How to add a new source DB or sink adapter |
| [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) | Community standards |

---

## Contributing

We welcome contributions, especially:

- **New source database adapters** (Oracle, Db2, MariaDB, CockroachDB, …)
- **New sink adapters** (BigQuery, Snowflake, Iceberg, Delta)
- **Connector config improvements** (gap recovery, retention, observability)
- **Documentation** (translations, deployment playbooks, real-world case studies)
- **Test data and reproducible demos** (Docker Compose fixtures for new DBs)

Before sending a pull request, please read [`CONTRIBUTING.md`](CONTRIBUTING.md). The project is small enough that you can ask questions by opening an issue.

---

## Community

- **Issues** — bug reports, feature requests, “does this work for X?” questions
- **Discussions** — architecture debates, design proposals, RFCs
- **Security** — please see [`SECURITY.md`](SECURITY.md) (TBD) for responsible disclosure

---

## Roadmap

- [x] Multi-source RDBMS CDC (PG, MySQL, SQL Server)
- [x] ClickHouse sink with deduplication-safe storage
- [x] Production troubleshooting guide
- [ ] MariaDB adapter validation
- [ ] Oracle adapter docs (license-aware)
- [ ] Db2 adapter research
- [ ] BigQuery sink (experimental)
- [ ] Multi-host production compose (Kafka cluster, CH cluster)
- [ ] Grafana dashboard JSON for lag / throughput
- [ ] Prometheus scrape config
- [ ] ClickHouse backup to S3 script
- [ ] End-to-end test harness with sample DBs in CI

See [`docs/open-source-vision.md`](docs/open-source-vision.md) for the long-term direction.

---

## License

Released under the **MIT License**. See [`LICENSE`](LICENSE).

You are free to use this in personal, commercial, and educational projects, including forks, as long as the copyright and license notice are preserved.
