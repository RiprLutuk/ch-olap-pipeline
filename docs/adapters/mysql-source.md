# MySQL Source Adapter

> **Status: 🟢 demo-ready**

## Overview

The MySQL adapter uses [Debezium MySQL Connector](https://debezium.io/documentation/reference/stable/connectors/mysql.html).

## Required Database Setup

1. **Enable Row-level Binlog**
   Modify `my.cnf`:
   ```ini
   [mysqld]
   server-id         = 223344
   log_bin           = mysql-bin
   binlog_format     = ROW
   binlog_row_image  = FULL
   expire_logs_days  = 10
   ```
   Restart MySQL.

2. **Create CDC User**
   ```sql
   CREATE USER 'cdc_user'@'%' IDENTIFIED BY 'secret';
   GRANT SELECT, RELOAD, SHOW DATABASES, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'cdc_user'@'%';
   FLUSH PRIVILEGES;
   ```

## Connector Config

File: `deploy/kafka-debezium/connectors/mysql-source.json`
