# SQL Server Source Adapter

> **Status: 🟢 demo-ready**

## Overview

The SQL Server adapter uses [Debezium SQL Server Connector](https://debezium.io/documentation/reference/stable/connectors/sqlserver.html). Requires Standard or Enterprise Edition (Web/Express do not support native CDC).

## Required Database Setup

1. **Enable CDC at Database Level**
   ```sql
   USE mydb;
   EXEC sys.sp_cdc_enable_db;
   ```

2. **Enable CDC at Table Level** (Required per table)
   ```sql
   EXEC sys.sp_cdc_enable_table
     @source_schema = N'dbo',
     @source_name   = N'orders',
     @role_name     = NULL;
   ```

## Connector Config

File: `deploy/kafka-debezium/connectors/sqlserver-source.json`

## Known Limitations
- SQL Server Agent must be running (it powers the capture jobs).
