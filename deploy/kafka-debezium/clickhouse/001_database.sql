CREATE DATABASE IF NOT EXISTS analytics;

CREATE USER IF NOT EXISTS analytics IDENTIFIED WITH plaintext_password BY 'change-me';
GRANT ALL ON analytics.* TO analytics;
