-- ClickHouse OLAP schema
-- Receives Debezium topics from Postgres + MySQL via Kafka Engine,
-- funnels them through Materialized Views into ReplacingMergeTree tables.

CREATE DATABASE IF NOT EXISTS shop_analytics;

-- ============================================================
-- Postgres Kafka source tables
-- ============================================================

CREATE TABLE IF NOT EXISTS shop_analytics.kafka_pg_customers
(
    id         Int64,
    name       String,
    email      String,
    city       String,
    created_at String,
    updated_at String,
    __deleted  String
)
ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'kafka:29092',
    kafka_topic_list = 'shop.public.customers',
    kafka_group_name = 'clickhouse-pg-customers',
    kafka_format = 'JSONEachRow',
    kafka_num_consumers = 1,
    kafka_skip_broken_messages = 100;

CREATE TABLE IF NOT EXISTS shop_analytics.kafka_pg_orders
(
    id           Int64,
    customer_id  Int64,
    status       String,
    total_amount Float64,
    created_at   String,
    updated_at   String,
    __deleted    String
)
ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'kafka:29092',
    kafka_topic_list = 'shop.public.orders',
    kafka_group_name = 'clickhouse-pg-orders',
    kafka_format = 'JSONEachRow',
    kafka_num_consumers = 1,
    kafka_skip_broken_messages = 100;

-- ============================================================
-- MySQL Kafka source tables
-- ============================================================

CREATE TABLE IF NOT EXISTS shop_analytics.kafka_mysql_orders
(
    id           Int64,
    customer_id  Int64,
    status       String,
    total_amount Float64,
    created_at   String,
    updated_at   String,
    __deleted    String
)
ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'kafka:29092',
    kafka_topic_list = 'shop_mysql.shop.orders',
    kafka_group_name = 'clickhouse-mysql-orders',
    kafka_format = 'JSONEachRow',
    kafka_num_consumers = 1,
    kafka_skip_broken_messages = 100;

-- ============================================================
-- Target tables (ReplacingMergeTree for upsert semantics)
-- ============================================================

CREATE TABLE IF NOT EXISTS shop_analytics.pg_customers
(
    id         Int64,
    name       String,
    email      String,
    city       String,
    created_at DateTime64(3),
    updated_at DateTime64(3),
    is_deleted UInt8
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY id;

CREATE TABLE IF NOT EXISTS shop_analytics.pg_orders
(
    id           Int64,
    customer_id  Int64,
    status       String,
    total_amount Decimal(14, 2),
    created_at   DateTime64(3),
    updated_at   DateTime64(3),
    is_deleted   UInt8
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (created_at, id);

CREATE TABLE IF NOT EXISTS shop_analytics.mysql_orders
(
    id           Int64,
    customer_id  Int64,
    status       String,
    total_amount Decimal(14, 2),
    created_at   DateTime64(3),
    updated_at   DateTime64(3),
    is_deleted   UInt8
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (created_at, id);

-- ============================================================
-- Order events (append-only history, no dedup → funnel-friendly)
-- ============================================================

CREATE TABLE IF NOT EXISTS shop_analytics.pg_orders_events
(
    id           Int64,
    customer_id  Int64,
    status       String,
    total_amount Decimal(14, 2),
    created_at   DateTime64(3),
    updated_at   DateTime64(3),
    is_deleted   UInt8
)
ENGINE = MergeTree
ORDER BY (id, updated_at);

-- ============================================================
-- Materialized Views
-- ============================================================

CREATE MATERIALIZED VIEW IF NOT EXISTS shop_analytics.mv_pg_customers
TO shop_analytics.pg_customers AS
SELECT
    id,
    name,
    email,
    city,
    parseDateTime64BestEffortOrZero(created_at, 3) AS created_at,
    parseDateTime64BestEffortOrZero(updated_at, 3) AS updated_at,
    toUInt8(__deleted = 'true') AS is_deleted
FROM shop_analytics.kafka_pg_customers;

CREATE MATERIALIZED VIEW IF NOT EXISTS shop_analytics.mv_pg_orders
TO shop_analytics.pg_orders AS
SELECT
    id,
    customer_id,
    status,
    toDecimal64(total_amount, 2) AS total_amount,
    parseDateTime64BestEffortOrZero(created_at, 3) AS created_at,
    parseDateTime64BestEffortOrZero(updated_at, 3) AS updated_at,
    toUInt8(__deleted = 'true') AS is_deleted
FROM shop_analytics.kafka_pg_orders;

CREATE MATERIALIZED VIEW IF NOT EXISTS shop_analytics.mv_pg_orders_events
TO shop_analytics.pg_orders_events AS
SELECT
    id,
    customer_id,
    status,
    toDecimal64(total_amount, 2) AS total_amount,
    parseDateTime64BestEffortOrZero(created_at, 3) AS created_at,
    parseDateTime64BestEffortOrZero(updated_at, 3) AS updated_at,
    toUInt8(__deleted = 'true') AS is_deleted
FROM shop_analytics.kafka_pg_orders;

CREATE MATERIALIZED VIEW IF NOT EXISTS shop_analytics.mv_mysql_orders
TO shop_analytics.mysql_orders AS
SELECT
    id,
    customer_id,
    status,
    toDecimal64(total_amount, 2) AS total_amount,
    parseDateTime64BestEffortOrZero(created_at, 3) AS created_at,
    parseDateTime64BestEffortOrZero(updated_at, 3) AS updated_at,
    toUInt8(__deleted = 'true') AS is_deleted
FROM shop_analytics.kafka_mysql_orders;
