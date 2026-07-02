CREATE TABLE IF NOT EXISTS analytics.orders_kafka
(
    raw_message String
)
ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'kafka:9092',
    kafka_topic_list = 'oltpdemo.public.orders',
    kafka_group_name = 'ch-orders-consumer',
    kafka_format = 'JSONAsString',
    kafka_num_consumers = 1;
