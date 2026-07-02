# PostgreSQL Source Adapter

> **Status: 🟢 demo-ready**

## Overview

The PostgreSQL adapter uses [Debezium PostgreSQL Connector](https://debezium.io/documentation/reference/stable/connectors/postgresql.html) using the `pgoutput` plugin (native to Postgres 10+).

## Required Database Setup

1. **Enable Logical Replication**
   Modify `postgresql.conf`:
   ```ini
   wal_level = logical
   max_wal_senders = 4
   max_replication_slots = 4
   ```
   Restart PostgreSQL.

2. **Create User & Permissions**
   ```sql
   CREATE ROLE cdc_user REPLICATION LOGIN PASSWORD 'secret';
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO cdc_user;
   ```

## Connector Config

File: `deploy/kafka-debezium/connectors/postgres-source.json`

## Known Limitations

- **Toast columns:** Unchanged TOAST columns won't be sent in UPDATE events.
- **DDL:** Debezium tracks schema changes, but dropping tables while CDC is running can cause failures if replication slots aren't managed.
