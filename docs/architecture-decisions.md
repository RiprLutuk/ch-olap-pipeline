# Architecture Decision Records

This document captures the **why** behind major decisions in `ch-olap-pipeline`. Future contributors should read this first.

---

## ADR-001 — Real CDC instead of polling

**Status:** Accepted

**Context:** Original reference project used polling (`SELECT ... WHERE updated_at > last_run`). In heterogeneous OLTP environments with DELETE-heavy workloads, polling is expensive, lossy, and operationally painful.

**Decision:** Use **Change Data Capture** via Debezium source connectors reading native database logs (WAL for PostgreSQL, binlog for MySQL, CDC job for SQL Server).

**Consequences:**

- Pro: low latency (< 1s per event), DELETE detection, low source DB overhead
- Pro: replayable event log in Kafka
- Con: more components (Kafka, Kafka Connect, Debezium)
- Con: source DB needs CDC-aware config (replication slots, binlog format, CDC jobs)

---

## ADR-002 — Apache Kafka as the event bus

**Status:** Accepted

**Context:** Need a durable, replayable, partitioned event bus between Debezium and the analytical sink.

**Decision:** Use Apache Kafka in **KRaft mode** (no ZooKeeper). Single-node for dev, multi-node for production.

**Consequences:**

- Pro: industry standard, tooling-rich, schema registry compatible
- Pro: KRaft removes ZK operational complexity
- Con: JVM footprint (~1-2GB minimum), not suitable for very small VPS

---

## ADR-003 — ClickHouse as primary sink

**Status:** Accepted

**Context:** Need a column-oriented analytical store with high-throughput ingestion and low-latency query. Alternatives: BigQuery (cloud-locked), Snowflake (cloud-locked, expensive), Iceberg (lake-format, not query engine), Druid (operationally heavy).

**Decision:** ClickHouse is the **first-class sink**. Other sinks are roadmap items, not v1 promises.

**Consequences:**

- Pro: native Kafka Engine, Materialized Views, ReplacingMergeTree for dedup
- Pro: self-hostable, on-prem friendly
- Con: less mature for non-Kafka ingestion paths

---

## ADR-004 — Single partition per topic for order guarantee

**Status:** Accepted

**Context:** CDC events for one table must be processed in commit order, otherwise UPDATE arrives before INSERT and the analytical state is wrong.

**Decision:** Force `topic.num.partitions=1` per source table. Scale horizontally by adding **more tables** or **more source databases**, not by splitting a single table's topic.

**Consequences:**

- Pro: simple ordering guarantee per row
- Con: limits per-topic throughput to one consumer
- Mitigation: tune `task.max.queue.size` and add more Connect workers for more parallel tables

---

## ADR-005 — `ReplacingMergeTree` + `FINAL` for deduplication

**Status:** Accepted

**Context:** Even with ordered topics, operators need idempotent ingestion and recovery from out-of-order arrivals during connector restarts or partition reassignment.

**Decision:** Use ClickHouse `ReplacingMergeTree(version)` with a version column derived from Debezium LSN / binlog position. User-facing queries use the `FINAL` modifier.

**Consequences:**

- Pro: idempotent, safe for retries and out-of-order events
- Pro: `FINAL` gives correct read semantics automatically
- Con: `FINAL` adds a per-query merge cost — acceptable for OLAP workloads, not ideal for high-frequency small reads

---

## ADR-006 — Modular connector configs, no hardcoded database list

**Status:** Accepted

**Context:** The project is positioned to grow beyond PostgreSQL, MySQL, and SQL Server. Hardcoding "supported databases" in core code would block extension.

**Decision:** Each source database is a **self-contained JSON config + docs** under `deploy/kafka-debezium/connectors/`. Adding a new source = dropping a new JSON + writing a section in the README. No code changes to the core pipeline.

**Consequences:**

- Pro: clean community contribution model
- Pro: each adapter has independent config knobs
- Con: per-DB setup steps live in markdown, not enforced by code — relies on contributors to keep docs accurate

---

## ADR-007 — Heartbeat intervals for replication slot keepalive

**Status:** Accepted

**Context:** PostgreSQL logical replication slots are **persistent** — if a connector dies and the slot is unused, WAL accumulates until disk is full. Operators routinely hit this in production.

**Decision:** Mandate `heartbeat.interval.ms=10000` in PG connector config. Document the **last-known LSN** check in troubleshooting.

**Consequences:**

- Pro: slot stays active, no silent disk fill
- Con: minor traffic overhead (one heartbeat per 10s)

---

## ADR-008 — No exposure of Kafka / Connect to public internet

**Status:** Accepted

**Context:** Kafka and Kafka Connect have no built-in auth in default config. Exposing them publicly = data exfiltration vector.

**Decision:** Default compose binds Kafka/Connect/ClickHouse to **localhost only**. Public access via reverse proxy is opt-in and requires explicit auth (SASL_SSL / HTTPS).

**Consequences:**

- Pro: secure default
- Con: operators must consciously add TLS / auth for remote access — documented in production checklist

---

## ADR-009 — Source DBs use low-privilege CDC users

**Status:** Accepted

**Context:** The CDC user is the only thing standing between the pipeline and a misconfiguration dropping a production table.

**Decision:** Document and require **per-DB CDC user** with:

- **PostgreSQL:** `REPLICATION` privilege + `SELECT` only on replicated tables
- **MySQL:** `REPLICATION SLAVE`, `REPLICATION CLIENT`, `SELECT` on replicated tables
- **SQL Server:** membership in `db_owner` for the source DB only (CDC requirements), or more restrictive if vendor allows

**Consequences:**

- Pro: blast radius limited if credentials leak
- Con: per-DB setup steps are not optional — documented in install guide

---

## ADR-010 — Documentation as a first-class deliverable

**Status:** Accepted

**Context:** The reference project this is built on shipped minimal docs. Operators adopting it hit the same wall repeatedly.

**Decision:** Every feature ships with:

- install steps
- verification commands
- troubleshooting section
- known limitations

Docs are reviewed with the same rigor as code.

**Consequences:**

- Pro: project is actually usable by people who did not build it
- Con: PRs feel heavier — but that is the right tradeoff
