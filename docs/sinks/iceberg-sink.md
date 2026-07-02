# Apache Iceberg Sink

Apache Iceberg is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> Apache Iceberg`

## Template

See `deploy/kafka-debezium/sinks/iceberg-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
