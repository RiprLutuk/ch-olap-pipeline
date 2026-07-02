# DuckDB Sink

DuckDB is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> DuckDB`

## Template

See `deploy/kafka-debezium/sinks/duckdb-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
