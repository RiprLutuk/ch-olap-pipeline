# System Overview

This document explains the project architecture in a more human-friendly way than the ADRs.

If the ADRs answer **why**, this page answers **what is here** and **how the parts fit together**.

## Simple summary

`ch-olap-pipeline` is a reference architecture for moving data changes from OLTP systems into ClickHouse using CDC.

In plain language:

1. a source database changes
2. a CDC connector captures the change
3. Kafka carries the event stream
4. ClickHouse consumes the stream
5. analytics tables become queryable

## Main data flow

```text
Source DB / API
      ↓
Debezium / Kafka Connect
      ↓
Kafka topics
      ↓
ClickHouse Kafka Engine
      ↓
Materialized Views
      ↓
Analytics tables
```

## Main components

### 1. Source systems
These are the operational systems where data is created or updated.

Examples:

- PostgreSQL
- MySQL
- SQL Server
- Oracle
- other future adapters

The source side should stay OLTP-safe. That means:

- low-privilege users
- replication-safe settings
- no destructive requirements on the source database

### 2. CDC capture layer
This layer watches changes from the source system.

For the primary path in this repo, that usually means:

- Debezium source connector
- Kafka Connect runtime

This is what turns database changes into structured events.

### 3. Kafka event bus
Kafka acts as the durable event stream between source capture and analytical storage.

Kafka is useful because it gives:

- replayability
- buffering
- decoupling between source and sink
- visibility into streamed events

But Kafka is also one of the heavier parts of the stack, especially on low-memory machines.

### 4. ClickHouse sink
ClickHouse is the first-class analytical target in this repository.

Why ClickHouse:

- very good ingestion throughput
- strong fit for analytical queries
- self-hostable
- practical for infra/data teams

The main pattern is:

- Kafka topic receives CDC events
- ClickHouse Kafka Engine reads them
- Materialized Views transform them
- final analytical tables store queryable data

## Deployment modes

## A. Learning / local lab mode
Best for:

- learning the architecture
- testing a single source
- validating configs

Typical environment:

- laptop
- local VM
- dedicated lab machine

This is the best place to run the full stack first.

## B. Small VPS mode
Best for:

- documentation hosting
- lightweight control plane
- demos that keep heavy components elsewhere

Important note:

A tiny VPS around 1GB RAM is usually **not** the right place for a full Kafka + Connect + Debezium + ClickHouse stack.

On small hosts, the realistic pattern is:

- keep heavy streaming components external
- use the VPS for light services, docs, dashboards, or app control surfaces

## C. Proper lab / production-like mode
Best for:

- realistic end-to-end tests
- multiple topics / larger datasets
- production-style learning

Typical target:

- 4–8GB RAM lab host or bigger
- separate services or separate nodes if scaling up

## What the repo promises today

This repository currently promises more strongly in these areas:

- architecture direction
- docs and decision records
- connector/sink structure
- community contribution path
- security-minded defaults

It currently promises less strongly in these areas:

- finished beginner UI
- plug-and-play product experience
- production-hardened implementation for every listed adapter

That distinction matters. The project is meant to be honest, not hype-driven.

## Current opinionated path

If you want the cleanest starting point, think of the project like this:

- start with **one source**: PostgreSQL, MySQL, or SQL Server
- use **Debezium + Kafka Connect** for CDC
- use **Kafka** as the event bus
- land the data in **ClickHouse**

That path should be the easiest one to understand, document, and improve first.

## How to read the docs

Recommended order:

1. `docs/index.md`
2. `docs/system-overview.md`
3. `docs/architecture-decisions.md`
4. `docs/supported-rdbms-matrix.md`
5. specific adapter docs under `docs/adapters/`
6. sink docs under `docs/sinks/`

## Future product direction

If the project becomes more beginner-friendly over time, the likely next step is:

- a simpler config format
- a guided CLI or setup wizard
- clearer quick-start flows for PostgreSQL/MySQL/SQL Server to ClickHouse

That is the practical path from **reference architecture** to **easy-to-use tool**.
