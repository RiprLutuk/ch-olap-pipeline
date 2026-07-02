# Open Source Vision — Universal CDC OLTP → OLAP Platform

## Mission

Build an open-source platform for engineers, DBAs, data platform teams, and backend/infra practitioners who need to replicate data changes from many OLTP RDBMS sources into analytical systems such as ClickHouse with **real CDC**, operational clarity, and production-grade troubleshooting.

The project should not be just a demo. It should become a **practical reference implementation** with:
- multi-database source support
- repeatable install paths
- observable pipelines
- clear failure recovery
- modular connectors
- open docs for infra/data people

## Open-source product principles

1. **Database-agnostic core** — PostgreSQL, MySQL, SQL Server are phase 1, but architecture must allow Oracle, MariaDB, IBM Db2, SQLite batch bridge, and others.
2. **Connector-driven design** — each source DB gets an isolated connector package/spec, not hardcoded assumptions inside one monolith.
3. **Production-first docs** — docs must explain scaling, gaps, recovery, retention, lag, and schema evolution.
4. **Ops-friendly defaults** — one-command local demo, then clear path to multi-host production.
5. **OLTP-safe** — low-privilege users, replication-safe config, no destructive requirements on source DBs.
6. **Extensible sinks** — ClickHouse first, but architecture should allow future sinks (BigQuery, Snowflake, Iceberg, Delta, Postgres warehouse).

## Target audience

- DBAs
- data engineers
- platform engineers
- backend / infra engineers
- teams modernizing from cron polling ETL to CDC streaming
- multi-database enterprise environments

## Positioning

This project sits between:
- lightweight DIY scripts that do polling and break under scale
- heavyweight enterprise CDC tools that are expensive and opaque
- single-database-focused replication tools that do not fit heterogeneous environments

It should feel like:
- easy enough for a small team to run
- rigorous enough for production learning
- modular enough to extend to new RDBMS
