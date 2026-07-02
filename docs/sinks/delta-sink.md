# Delta Lake Sink

Delta Lake is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> Delta Lake`

## Template

See `deploy/kafka-debezium/sinks/delta-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
