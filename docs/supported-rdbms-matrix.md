# Supported RDBMS Matrix

## Philosophy

Support should be declared in stages, not hand-waved. Each database source needs clear status:
- `demo-ready`
- `beta`
- `planned`
- `research`

## Source support matrix

| Source DB | CDC mechanism | Likely adapter path | Status target | Notes |
|---|---|---|---|---|
| PostgreSQL | WAL logical replication | Debezium PG | demo-ready | Strong first-class support |
| MySQL | binlog row-based | Debezium MySQL | demo-ready | Includes Aurora MySQL-compatible |
| MariaDB | binlog row-based | Debezium/MySQL-compatible validation | beta | Needs compatibility test matrix |
| SQL Server | native CDC | Debezium SQL Server | demo-ready | Table-level CDC enablement required |
| Oracle | LogMiner / XStream | Debezium Oracle | beta | Licensing/ops caveats must be documented |
| IBM Db2 | transaction log capture | Debezium Db2 | planned | Enterprise use-case, lower community access |
| Vitess / PlanetScale-like MySQL | binlog via compatible path | custom/validated MySQL path | planned | Validate GTID/binlog semantics |
| YugabyteDB (PG wire) | logical replication variant | PG-compatible adapter | research | Compatibility unknown |
| CockroachDB | changefeeds / CDC | native adapter | planned | Likely not Debezium-first |
| SQLite | no native CDC log stream | batch bridge / WAL parsing | research | Not true universal CDC without compromise |
| Firebird | trigger/log based bridge | custom adapter | research | Community contribution candidate |
| SAP HANA | log-based / SDI path | custom/enterprise adapter | research | May need separate plugin model |

## Sink matrix

| Sink | Primary target | Status |
|---|---|---|
| ClickHouse | first-class | active |
| BigQuery | future | planned |
| Snowflake | future | planned |
| Apache Iceberg | future | planned |
| Delta Lake | future | planned |
| PostgreSQL warehouse | future | planned |

## Support policy

A source DB is only considered supported when it has:
1. setup guide
2. required DB-side config documented
3. connector template
4. sample table flow verified end-to-end
5. troubleshooting section
6. known limitations list

## Recommendation

For launch messaging, claim:
- **First-class:** PostgreSQL, MySQL, SQL Server
- **Roadmap:** MariaDB, Oracle, Db2, CockroachDB
- **Research/Community:** SQLite, Firebird, HANA, YugabyteDB

Do **not** market “supports every RDBMS” until each adapter has verified docs/tests. Better wording:

> Built to become a universal RDBMS CDC platform, with first-class support starting from PostgreSQL, MySQL, and SQL Server, then expanding through modular source adapters.
