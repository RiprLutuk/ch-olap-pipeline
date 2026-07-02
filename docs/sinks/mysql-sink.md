# MySQL Warehouse Sink

MySQL Warehouse is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> MySQL Warehouse`

## Template

See `deploy/kafka-debezium/sinks/mysql-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
