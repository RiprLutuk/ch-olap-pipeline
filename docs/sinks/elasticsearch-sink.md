# Elasticsearch Sink

Elasticsearch is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> Elasticsearch`

## Template

See `deploy/kafka-debezium/sinks/elasticsearch-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
