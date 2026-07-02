# PostgreSQL Warehouse Sink

PostgreSQL Warehouse is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> PostgreSQL Warehouse`

## Template

See `deploy/kafka-debezium/sinks/postgres-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
