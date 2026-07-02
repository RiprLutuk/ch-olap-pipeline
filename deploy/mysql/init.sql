-- MySQL OLTP schema for CDC demo
-- Requires binlog-format=ROW + server-id + gtid-mode=ON.

CREATE DATABASE IF NOT EXISTS shop;

USE shop;

CREATE TABLE IF NOT EXISTS customers (
    id          BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    email       VARCHAR(200) NOT NULL UNIQUE,
    city        VARCHAR(100) NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS products (
    id          BIGINT         NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(200)   NOT NULL,
    category    VARCHAR(100)   NOT NULL,
    price       DECIMAL(12, 2) NOT NULL,
    stock       INT            NOT NULL DEFAULT 0,
    created_at  TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS orders (
    id            BIGINT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    customer_id   BIGINT          NOT NULL,
    status        VARCHAR(20)     NOT NULL,
    total_amount  DECIMAL(14, 2)  NOT NULL,
    created_at    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_orders_customer (customer_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_items (
    id          BIGINT         NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_id    BIGINT         NOT NULL,
    product_id  BIGINT         NOT NULL,
    quantity    INT            NOT NULL,
    unit_price  DECIMAL(12, 2) NOT NULL,
    KEY idx_items_order (order_id)
) ENGINE=InnoDB;

-- Seed products.
INSERT IGNORE INTO products (id, name, category, price, stock) VALUES
    (1, 'Kopi Arabica Gayo 250g',  'Beverages',   85000,  200),
    (2, 'Teh Hijau Premium 100g',  'Beverages',   45000,  150),
    (3, 'Keyboard Mechanical RGB', 'Electronics', 1250000, 30),
    (4, 'Mouse Wireless Ergonomis','Electronics', 350000,  80),
    (5, 'Notebook A5 Hardcover',   'Stationery',  55000,  500),
    (6, 'Pulpen Gel 0.5mm',        'Stationery',  12000,  1000),
    (7, 'Tas Ransel Kanvas',       'Fashion',     275000,  60),
    (8, 'Sepatu Sneakers Putih',   'Fashion',     650000,  40);
