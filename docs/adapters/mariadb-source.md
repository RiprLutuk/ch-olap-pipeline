---
title: MariaDB Source Adapter
tags: [source, cdc, mariadb, scaffold]
status: GA scaffold
---

# MariaDB Source Adapter

MariaDB coverage is included so the project can grow beyond the default PostgreSQL/MySQL/SQL Server path.

## Mechanism

- Source-specific CDC, polling, or log bridge
- Kafka Connect compatible source contract
- Events flow into Kafka and then ClickHouse or another sink

## Connector template

See `deploy/kafka-debezium/connectors/mariadb-source.json.example`.

## Status

| Dimension | State |
|---|---|
| Coverage | GA scaffold |
| Template | available |
| Production validation | contribution welcome |

This adapter is documented with a connector contract and template. Production hardening depends on the underlying database and connector implementation.
