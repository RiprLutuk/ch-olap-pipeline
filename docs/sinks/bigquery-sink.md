# BigQuery Sink

BigQuery is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> BigQuery`

## Template

See `deploy/kafka-debezium/sinks/bigquery-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
