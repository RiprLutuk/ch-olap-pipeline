# ClickHouse Sink

ClickHouse is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> ClickHouse`

## Template

See `deploy/kafka-debezium/sinks/clickhouse-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
