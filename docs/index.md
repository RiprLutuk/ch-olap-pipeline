# ch-olap-pipeline

> Universal OLTP → OLAP CDC pipeline with Debezium, Kafka, and ClickHouse.

`ch-olap-pipeline` is an open-source reference platform for moving data changes from heterogeneous OLTP databases into analytical storage using real Change Data Capture.

## What this project is for

- DBAs who need a transparent CDC reference implementation
- Data engineers who want a practical Debezium + Kafka + ClickHouse stack
- Platform engineers who need observable, reproducible data pipelines
- Backend/infra teams who deal with mixed RDBMS environments

## Core stack

```text
PostgreSQL / MySQL / SQL Server / Oracle / ...
        ↓
Debezium Kafka Connect
        ↓
Apache Kafka / KRaft
        ↓
ClickHouse Kafka Engine + Materialized Views
        ↓
Analytics tables
```

## Start here

- [Open Source Vision](open-source-vision.md)
- [Supported RDBMS Matrix](supported-rdbms-matrix.md)
- [Architecture Decisions](architecture-decisions.md)
- [Oracle Source Adapter](adapters/oracle-source.md)

## Repository

GitHub: [RiprLutuk/ch-olap-pipeline](https://github.com/RiprLutuk/ch-olap-pipeline)

## License

MIT
