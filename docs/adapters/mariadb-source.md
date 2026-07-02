# MariaDB Source Adapter

> **Status: 🟡 beta** — Needs extensive validation due to GTID differences vs MySQL.

## Overview

MariaDB is supported via the Debezium MySQL connector (up to MariaDB 10.x).

## Required Setup

See the [MySQL Source Adapter](mysql-source.md) setup. Binlog config is identical.

## Connector Config

File: `deploy/kafka-debezium/connectors/mariadb-source.json.example`
