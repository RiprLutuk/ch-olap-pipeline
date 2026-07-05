# CH OLAP Pipeline

> Community OLTP → OLAP CDC pipeline for DBA, infra, backend, and data-platform engineers.

`ch-olap-pipeline` is an open-source reference architecture for moving data changes from heterogeneous operational databases into analytical storage using **real Change Data Capture**, **streaming transport**, and **production-friendly sink patterns**.

---

## Why this project exists

Most OLTP → OLAP examples stop too early:
- they only support one source database
- they skip operational concerns
- they ignore schema evolution and failure handling
- they look good in diagrams but break in mixed real-world environments

`ch-olap-pipeline` is designed as a **practical engineering reference**, not a toy demo.

It is built for teams that need to handle:
- mixed database estates across business units
- CDC replication with clear operational boundaries
- scalable ingestion into ClickHouse and other analytical sinks
- observable, reproducible, community-friendly pipeline design

---

## Core architecture

```text
PostgreSQL / MySQL / MariaDB / SQL Server / Oracle / Db2 / MongoDB / ...
                                 ↓
                    Source Adapters / CDC Connectors
                                 ↓
                    Debezium + Kafka Connect / Native Ingest
                                 ↓
                         Apache Kafka / KRaft
                                 ↓
                Sink Consumers / Kafka Engine / Stream Processors
                                 ↓
        ClickHouse / Iceberg / Delta / BigQuery / Snowflake / others
```

---

## What makes it useful

### Multi-source by design
The project is shaped for heterogeneous environments, not single-engine comfort zones. PostgreSQL, MySQL, MariaDB, SQL Server, Oracle, and additional engines are treated as first-class citizens in the architecture direction.

### Built for operations
This project does not only care about ingestion. It also cares about:
- connector boundaries
- retries and failure domains
- sink behavior
- schema evolution
- observability
- deployment realism

### ClickHouse-forward, not ClickHouse-only
ClickHouse is the primary sink direction today, but the architecture is intentionally broader so the project can serve as a reusable reference for other analytical targets too.

### Community open-source posture
The goal is not just to publish code. The goal is to build a **real OSS project** that DBA, infra, and data-platform engineers can understand, adopt, extend, and contribute to.

---

## Start here

### Understand the direction
- [Open Source Vision](open-source-vision.md)
- [System Overview](system-overview.md)
- [Architecture Decisions](architecture-decisions.md)

### Explore source support
- [Supported RDBMS Matrix](supported-rdbms-matrix.md)
- [PostgreSQL Source Adapter](adapters/postgres-source.md)
- [MySQL Source Adapter](adapters/mysql-source.md)
- [SQL Server Source Adapter](adapters/sqlserver-source.md)
- [Oracle Source Adapter](adapters/oracle-source.md)

### Explore sink targets
- [Sinks Matrix](sinks-matrix.md)
- [ClickHouse Sink](sinks/clickhouse-sink.md)
- [Kafka Sink](sinks/kafka-sink.md)
- [Iceberg Sink](sinks/iceberg-sink.md)
- [BigQuery Sink](sinks/bigquery-sink.md)

---

## Who this is for

This documentation is especially useful for:
- **DBAs** who need a transparent CDC reference they can reason about
- **Infrastructure engineers** who need reliable service boundaries and deployment clarity
- **Backend engineers** who need event-driven data flow without magical abstractions
- **Data-platform engineers** who need realistic ingestion patterns for analytical systems
- **Technical leads** who want a vendor-neutral reference for multi-database OLTP → OLAP design

---

## Design principles

1. **Operational clarity over buzzwords**  
   Every component should have a clear job, failure domain, and scaling story.

2. **Heterogeneous source support matters**  
   Real organizations do not run one clean database engine forever.

3. **Streaming should stay understandable**  
   The pipeline should be powerful without becoming unreadable.

4. **Open-source quality is part of the architecture**  
   Good docs, contribution flow, decisions, and reproducibility are part of the system.

---

## Project status

`ch-olap-pipeline` is evolving toward a broader, production-aware open-source reference platform for:
- multi-database CDC ingestion
- Kafka-centered transport
- ClickHouse-first analytics delivery
- extensible sink architecture
- clearer operational and deployment guidance

---

## Repository

- GitHub: [RiprLutuk/ch-olap-pipeline](https://github.com/RiprLutuk/ch-olap-pipeline)

## License

MIT
