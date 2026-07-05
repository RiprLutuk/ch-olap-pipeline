# Supported RDBMS Matrix

The goal of this project is to provide reference CDC implementations for a wide variety of operational databases.

The matrix below tracks the current status of each source adapter.

---

## Status definitions

- <span style="color:#059669; font-weight:bold;">● GA</span> : Fully tested, documentation complete, ready for production reference.
- <span style="color:#d97706; font-weight:bold;">● Beta</span> : Works in lab, connector template provided, edge cases might need tuning.
- <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> : Placeholder created. PRs welcome to build out the implementation.

---

## Source adapters

| Database Engine | Status | Adapter Documentation | Connector Template |
|:---|:---:|:---|:---|
| **PostgreSQL** | <span style="color:#059669; font-weight:bold;">● GA</span> | [postgres-source.md](adapters/postgres-source.md) | `postgres-source.json.example` |
| **MySQL** | <span style="color:#059669; font-weight:bold;">● GA</span> | [mysql-source.md](adapters/mysql-source.md) | `mysql-source.json.example` |
| **SQL Server** | <span style="color:#059669; font-weight:bold;">● GA</span> | [sqlserver-source.md](adapters/sqlserver-source.md) | `sqlserver-source.json.example` |
| **Oracle** | <span style="color:#d97706; font-weight:bold;">● Beta</span> | [oracle-source.md](adapters/oracle-source.md) | `oracle-source.json.example` |
| **MariaDB** | <span style="color:#d97706; font-weight:bold;">● Beta</span> | [mariadb-source.md](adapters/mariadb-source.md) | `mariadb-source.json.example` |
| **IBM Db2** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [db2-source.md](adapters/db2-source.md) | `db2-source.json.example` |
| **CockroachDB** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [cockroachdb-source.md](adapters/cockroachdb-source.md) | `cockroachdb-source.json.example` |
| **YugabyteDB** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [yugabyte-source.md](adapters/yugabyte-source.md) | `yugabyte-source.json.example` |
| **Vitess** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [vitess-source.md](adapters/vitess-source.md) | `vitess-source.json.example` |
| **MongoDB** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [mongodb-source.md](adapters/mongodb-source.md) | `mongodb-source.json.example` |
| **SQLite** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [sqlite-source.md](adapters/sqlite-source.md) | `sqlite-source.json.example` |
| **Firebird** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [firebird-source.md](adapters/firebird-source.md) | `firebird-source.json.example` |
| **SAP HANA** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [sap-hana-source.md](adapters/sap-hana-source.md) | `sap-hana-source.json.example` |
| **Kafka-native** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [kafka-source.md](adapters/kafka-source.md) | `kafka-source.json.example` |
| **REST / Webhook** | <span style="color:#64748b; font-weight:bold;">○ Scaffold</span> | [rest-source.md](adapters/rest-source.md) | `rest-source.json.example` |

---

## Contributing a new source

If you want to move a source from `Scaffold` to `Beta` or `GA`:

1. Check the adapter documentation file linked above.
2. Provide the Debezium (or alternative) configuration template.
3. Document the specific database prerequisites (e.g., WAL levels, user grants, binlog settings).
4. Submit a Pull Request.