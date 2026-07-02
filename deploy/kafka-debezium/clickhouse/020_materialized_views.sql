CREATE TABLE IF NOT EXISTS analytics.orders_raw
(
    ingested_at DateTime DEFAULT now(),
    raw_message String
)
ENGINE = MergeTree
ORDER BY ingested_at;

CREATE MATERIALIZED VIEW IF NOT EXISTS analytics.mv_orders_raw
TO analytics.orders_raw
AS
SELECT now() AS ingested_at, raw_message
FROM analytics.orders_kafka;
