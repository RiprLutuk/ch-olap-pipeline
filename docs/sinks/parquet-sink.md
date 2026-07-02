# Parquet Files Sink

Parquet Files is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> Parquet Files`

## Template

See `deploy/kafka-debezium/sinks/parquet-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
