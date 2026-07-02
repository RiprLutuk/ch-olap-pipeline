# Webhook Sink

Webhook is documented as a sink target for the OLTP to OLAP pipeline.

## Flow

`Source DB -> Kafka Connect -> Kafka topics -> Webhook`

## Template

See `deploy/kafka-debezium/sinks/webhook-sink.json.example`.

## Status

| Dimension | State |
|---|---|
| Template | available |
| Target type | sink |
| Production validation | contribution welcome |
