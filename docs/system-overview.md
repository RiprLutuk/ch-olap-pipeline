# System Overview

This document explains the project architecture from a high-level operational perspective. If the ADRs answer **why**, this page answers **what is here**, **how the pieces fit together**, and **what deployment shape makes sense in the real world**.

---

## The 10,000-foot view

`ch-olap-pipeline` is a reference architecture for capturing data changes from operational systems (OLTP) and delivering them reliably into analytical storage (OLAP) using CDC and streaming.

In plain language:

1. **A source database changes** through inserts, updates, or deletes.
2. **A CDC connector captures the change** from the database log.
3. **Kafka carries and buffers the event stream**.
4. **ClickHouse consumes the stream**.
5. **Analytics tables become queryable** without putting analytical load on the source database.

---

## Architectural flow

```text
┌─────────────────┐       ┌──────────────────────┐       ┌─────────────────┐
│                 │       │                      │       │                 │
│  Source RDBMS   ├──────►│ Debezium / Connect   ├──────►│  Apache Kafka   │
│  or API Source  │       │   CDC Capture Layer  │       │                 │
│                 │       │                      │       │                 │
└─────────────────┘       └──────────────────────┘       └────────┬────────┘
                                                                   │
                                                                   │
                                              ┌────────────────────┴────────────────────┐
                                              │                                         │
                                              ▼                                         ▼
                                   ┌─────────────────┐                        ┌─────────────────┐
                                   │   ClickHouse    │                        │ Other Sinks /   │
                                   │ Kafka Engine +  │                        │ Future Targets  │
                                   │ Materialized    │                        │                 │
                                   │ Views           │                        │                 │
                                   └────────┬────────┘                        └─────────────────┘
                                            │
                                            ▼
                                   ┌─────────────────┐
                                   │ Analytics Tables│
                                   └─────────────────┘
```

---

## Layer breakdown

### 1. Source systems
These are the operational systems where data is created or updated.

Examples:
- PostgreSQL
- MySQL
- MariaDB
- SQL Server
- Oracle
- future adapters and API-based sources

The source side should stay OLTP-safe. That means:
- low-privilege users where possible
- replication-safe settings
- no destructive requirements on the source database
- no heavy polling as the primary ingestion strategy

### 2. CDC capture layer
This layer watches changes from the source system and turns them into structured events.

For the primary path in this repository, that usually means:
- Debezium source connector
- Kafka Connect runtime

This is where engine-specific transaction logs become portable event payloads.

### 3. Kafka event bus
Kafka acts as the durable event stream between source capture and analytical storage.

Kafka is useful because it gives:
- replayability
- buffering
- decoupling between source and sink
- visibility into streamed events
- safer recovery when downstream systems are unavailable

Kafka is also one of the heavier parts of the stack, especially on low-memory machines. That trade-off should be explicit.

### 4. ClickHouse sink
ClickHouse is the first-class analytical target in this repository.

Why ClickHouse:
- very strong ingestion throughput
- excellent fit for analytical queries
- self-hostable
- practical for infra/data teams
- works well with Kafka-based ingestion patterns

The main pattern is:
- Kafka topic receives CDC events
- ClickHouse Kafka Engine reads them
- Materialized Views transform them
- final analytical tables store queryable data

### 5. Extensible sink direction
Although ClickHouse is the primary target, the project is intentionally broader.

The architecture is designed so teams can reason about additional sinks such as:
- Iceberg
- Delta Lake
- BigQuery
- Snowflake
- Redshift
- PostgreSQL or other operational mirrors

That keeps the project useful as a **reference platform**, not just a single-destination demo.

---

## Deployment modes

### A. Learning / local lab mode
Best for:
- learning the architecture
- testing a single source
- validating configs

Typical environment:
- laptop
- local VM
- dedicated lab machine

This is the best place to run the full stack first.

### B. Small VPS mode
Best for:
- documentation hosting
- lightweight control plane
- demos that keep heavy components elsewhere

Important note:

A tiny VPS around 1GB RAM is usually **not** the right place for a full Kafka + Connect + Debezium + ClickHouse stack.

On small hosts, the realistic pattern is:
- keep heavy streaming components external
- use the VPS for light services, docs, dashboards, or app control surfaces

### C. Proper lab / production-like mode
Best for:
- realistic end-to-end tests
- multiple topics and larger datasets
- production-style learning

Typical target:
- 4–8GB RAM lab host or bigger
- separate services or separate nodes if scaling up
- clearer network boundaries between source, stream, and analytics zones

---

## What the repo promises today

This repository currently promises more strongly in these areas:
- architecture direction
- docs and decision records
- connector and sink structure
- community contribution path
- security-minded defaults

It currently promises less strongly in these areas:
- finished beginner UI
- plug-and-play product experience
- production-hardened implementation for every listed adapter

That distinction matters. The project should be honest, not hype-driven.

---

## Current opinionated path

If you want the cleanest starting point, think of the project like this:
- start with **one source**: PostgreSQL, MySQL, SQL Server, or Oracle
- use **Debezium + Kafka Connect** for CDC
- use **Kafka** as the event bus
- land the data in **ClickHouse**

That path is the easiest one to understand, document, test, and improve first.

---

## How to read the docs

Recommended order:

1. `docs/index.md`
2. `docs/system-overview.md`
3. `docs/architecture-decisions.md`
4. `docs/supported-rdbms-matrix.md`
5. specific adapter docs under `docs/adapters/`
6. sink docs under `docs/sinks/`

---

## Future product direction

If the project becomes more beginner-friendly over time, the likely next step is:
- a simpler config format
- a guided CLI or setup wizard
- clearer quick-start flows for PostgreSQL / MySQL / SQL Server / Oracle → ClickHouse
- stronger validation around connector prerequisites
- better operational dashboards and troubleshooting guides

That is the practical path from **reference architecture** to **easy-to-use tool**.
