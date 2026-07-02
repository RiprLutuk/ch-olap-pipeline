---
title: IBM Db2 Source Adapter
tags: [source, cdc, db2, scaffold]
status: scaffold
---

# IBM Db2 Source Adapter

IBM Db2 coverage is included so the project can grow beyond the default PostgreSQL/MySQL/SQL Server path.

## Mechanism

- Source-specific CDC, polling, or log bridge
- Kafka Connect compatible source contract
- Events flow into Kafka and then ClickHouse or another sink

## Connector template

See `deploy/kafka-debezium/connectors/db2-source.json.example`.

## Status

| Dimension | State |
|---|---|
| Coverage | scaffold |
| Template | available |
| Production validation | contribution welcome |

This adapter is documented with a connector contract and template. Production hardening depends on the underlying database and connector implementation.
