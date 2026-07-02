---
title: SAP HANA Source Adapter
tags: [source, cdc, sap-hana, scaffold]
status: scaffold
---

# SAP HANA Source Adapter

SAP HANA coverage is included so the project can grow beyond the default PostgreSQL/MySQL/SQL Server path.

## Mechanism

- Source-specific CDC, polling, or log bridge
- Kafka Connect compatible source contract
- Events flow into Kafka and then ClickHouse or another sink

## Connector template

See `deploy/kafka-debezium/connectors/sap-hana-source.json.example`.

## Status

| Dimension | State |
|---|---|
| Coverage | scaffold |
| Template | available |
| Production validation | contribution welcome |

This adapter is documented with a connector contract and template. Production hardening depends on the underlying database and connector implementation.
