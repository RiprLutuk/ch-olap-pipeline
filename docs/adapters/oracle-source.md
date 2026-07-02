# Oracle Source Adapter

> **Status: 🟡 beta** — config template, DB-side setup, and known limitations are documented. Not yet verified end-to-end in CI.

---

## Overview

Oracle CDC is available through the [Debezium Oracle Connector](https://debezium.io/documentation/reference/stable/connectors/oracle.html), which reads directly from Oracle's **LogMiner** (or optionally **XStream**).

This is one of the most complex adapters in the pipeline due to Oracle's licensing model, redo log configuration, and supplemental logging requirements — all of which this guide covers.

---

## Oracle version support

| Version | LogMiner | XStream | Notes |
|---|---|---|---|
| Oracle 11g | ✅ | ⚠️ limited | Requires supplemental logging |
| Oracle 12c | ✅ | ✅ | Recommended minimum |
| Oracle 19c | ✅ | ✅ | Recommended LTS version |
| Oracle 21c | ✅ | ✅ | |
| Oracle 23c (free) | ✅ | ⚠️ not confirmed | Community report only |

---

## Licensing caveat

> ⚠️ **Important:** LogMiner is available in all Oracle editions, including Standard Edition. XStream requires **Oracle GoldenGate license**. This project uses LogMiner by default.

---

## Required Oracle-side setup

### 1. Enable archive log mode

```sql
-- Run as SYSDBA
SHUTDOWN IMMEDIATE;
STARTUP MOUNT;
ALTER DATABASE ARCHIVELOG;
ALTER DATABASE OPEN;
```

Verify:
```sql
SELECT LOG_MODE FROM V$DATABASE;
-- Expected: ARCHIVELOG
```

### 2. Enable supplemental logging

```sql
-- Minimum: database-level supplemental log for all columns
ALTER DATABASE ADD SUPPLEMENTAL LOG DATA;
ALTER DATABASE ADD SUPPLEMENTAL LOG DATA (ALL) COLUMNS;
```

Or per-table (recommended for large databases):

```sql
ALTER TABLE SCHEMA.TABLE ADD SUPPLEMENTAL LOG DATA (ALL) COLUMNS;
```

### 3. Create a low-privilege CDC user

```sql
CREATE USER c##debezium IDENTIFIED BY your_password
  DEFAULT TABLESPACE users
  TEMPORARY TABLESPACE temp;

-- Required privileges
GRANT CREATE SESSION TO c##debezium;
GRANT SET CONTAINER TO c##debezium;
GRANT SELECT ON V_$DATABASE TO c##debezium;
GRANT FLASHBACK ANY TABLE TO c##debezium;
GRANT SELECT ANY TABLE TO c##debezium;
GRANT SELECT_CATALOG_ROLE TO c##debezium;
GRANT EXECUTE_CATALOG_ROLE TO c##debezium;
GRANT SELECT ANY TRANSACTION TO c##debezium;
GRANT LOGMINING TO c##debezium;
GRANT CREATE TABLE TO c##debezium;
GRANT LOCK ANY TABLE TO c##debezium;
GRANT CREATE SEQUENCE TO c##debezium;
GRANT EXECUTE ON DBMS_LOGMNR TO c##debezium;
GRANT EXECUTE ON DBMS_LOGMNR_D TO c##debezium;
GRANT SELECT ON V_$LOG TO c##debezium;
GRANT SELECT ON V_$LOG_HISTORY TO c##debezium;
GRANT SELECT ON V_$LOGMNR_LOGS TO c##debezium;
GRANT SELECT ON V_$LOGMNR_CONTENTS TO c##debezium;
GRANT SELECT ON V_$LOGMNR_PARAMETERS TO c##debezium;
GRANT SELECT ON V_$LOGFILE TO c##debezium;
GRANT SELECT ON V_$ARCHIVED_LOG TO c##debezium;
GRANT SELECT ON V_$ARCHIVE_DEST_STATUS TO c##debezium;
GRANT SELECT ON V_$TRANSACTION TO c##debezium;
```

> **Note:** For non-CDB (non-pluggable) Oracle, use a regular schema user without the `c##` prefix.

### 4. Verify redo log setup

```sql
SELECT GROUP#, MEMBERS, BYTES/1024/1024 MB, STATUS FROM V$LOG;
```

LogMiner needs at least **3 redo log groups** and ideally **200MB+ per group** for active pipelines.

---

## Connector config

File: [`connectors/oracle-source.json.example`](connectors/oracle-source.json.example)

Rename to `oracle-source.json` and set the environment variables in `.env`:

```bash
ORACLE_HOST=your-oracle-host
ORACLE_PORT=1521
ORACLE_USER=c##debezium
ORACLE_PASSWORD=your_password
ORACLE_SID=ORCL
ORACLE_PDB=ORCLPDB1   # leave empty for non-CDB
ORACLE_SCHEMA=APP
ORACLE_TABLE_INCLUDE=APP.ORDERS,APP.CUSTOMERS
TOPIC_PREFIX=oltpdemo
```

Then run:

```bash
set -a; source .env; set +a
./scripts/register-connectors.sh
```

---

## Verifying connector is running

```bash
# List all connectors
curl -s http://localhost:8083/connectors | jq .

# Check status of oracle connector
curl -s http://localhost:8083/connectors/oracle-source/status | jq .

# Expected
# "state": "RUNNING"
```

---

## Verifying events arrive in Kafka

```bash
# Start consuming the topic (table = APP.ORDERS)
docker exec ch-kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic oltpdemo.APP.APP.ORDERS \
  --from-beginning \
  --max-messages 5

# The topic name pattern is: {prefix}.{schema}.{schema}.{table}
# (Oracle includes schema in both positions)
```

---

## Known limitations

| Limitation | Detail |
|---|---|
| **Log retention** | Archive logs must be retained long enough for connector restarts. Default 24h via `log.mining.archive.log.hours` |
| **LOB columns** | LOB (CLOB, BLOB, NCLOB) not supported by default — use `lob.enabled=true` (Debezium 1.9+) |
| **RAC** | Oracle RAC (Real Application Clusters) needs extra config for redo log discovery |
| **Supplemental log gap** | If supplemental logging is enabled after tables have data, the connector must do an initial snapshot first |
| **Schema changes** | DDL changes (ALTER TABLE) require connector pause + restart in many scenarios |
| **JDBC driver** | Oracle JDBC driver (`ojdbc8.jar`) is **not bundled** in the Debezium image — must be mounted manually due to Oracle license restrictions |

---

## Adding JDBC driver to Connect image

Oracle's JDBC driver cannot be redistributed. You must add it yourself:

```bash
# Download from https://www.oracle.com/database/technologies/appdev/jdbc-downloads.html
# Then mount into Kafka Connect:
```

Add to `docker-compose.yml` under the `connect` service:

```yaml
volumes:
  - ./drivers/ojdbc8.jar:/kafka/libs/ojdbc8.jar:ro
```

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `ORA-01325: archive log mode must be enabled` | ARCHIVELOG not enabled | Run step 1 above |
| `Supplemental logging not configured` | Missing ALTER DATABASE | Run step 2 |
| `Access denied to V$LOGMNR_CONTENTS` | Missing privilege | Re-run GRANT statements |
| `No more data to read from socket` | Archive log already purged | Increase retention or restart connector |
| `ojdbc8.jar not found` | Driver not mounted | See JDBC driver section above |
| Events stale / connector appears running but no messages | Redo logs overwritten faster than consumed | Increase redo log group size, check `log.mining.archive.log.hours` |

---

## Contributing

If you have tested Oracle CDC in production and want to improve this adapter, please:

1. Open a GitHub Discussion in the **Source adapters** category
2. Share your environment (edition, version, CDB vs non-CDB)
3. Submit a PR with:
   - Updated connector config (if config changed)
   - Updated known limitations (if you hit something new)
   - Any verified workarounds

Your real-world experience makes this more useful for everyone.
