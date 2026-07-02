---
title: MySQL Source Adapter
tags: [source, cdc, mysql, scaffold]
status: GA
---

# MySQL Source Adapter

MySQL coverage is included so the project can grow beyond the default PostgreSQL/MySQL/SQL Server path.

## Mechanism

- Source-specific CDC, polling, or log bridge
- Kafka Connect compatible source contract
- Events flow into Kafka and then ClickHouse or another sink

## Connector template

See `deploy/kafka-debezium/connectors/mysql-source.json.example`.

## Status

| Dimension | State |
|---|---|
| Coverage | GA |
| Template | available |
| Production validation | contribution welcome |

This adapter is documented with a connector contract and template. Production hardening depends on the underlying database and connector implementation.
