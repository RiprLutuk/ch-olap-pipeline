# S3 Sink

S3 is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> S3`

## Template

See `deploy/kafka-debezium/sinks/s3-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
