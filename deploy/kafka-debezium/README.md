# Kafka + Debezium → ClickHouse — Full Architecture Guide

> **Target:** server terpisah **minimal 4GB RAM, ideal 8GB+**.  
> **JANGAN** deploy stack ini ke VPS DDAG 908MB.

---

## 1. Arsitektur CDC OLTP → OLAP

CDC = **Change Data Capture**. Bukan polling, bukan trigger. DB-nya sendiri yang ngomong "ada row baru, ada row yang berubah, ada yang dihapus" lewat log biner internal.

### 1.1 Kenapa gak langsung polling?

| Polling (kayak v1 native) | CDC (Kafka + Debezium) |
|---|---|
| SELECT ... WHERE updated_at > last_run tiap 5 detik | Stream event langsung dari WAL/binlog |
| Latency 1-5 detik batch | Latency < 1 detik per event |
| Beban DB naik proporsional ke row count | Beban DB konstan & minimal |
| Gak detect DELETE | Detect INSERT, UPDATE, DELETE semua |
| Susah scale ke 100+ table | Sekali register, jalan terus |

### 1.2 Sumber CDC per database

| DB Source | Mekanisme CDC | Plugin Debezium |
|---|---|---|
| **PostgreSQL** | Logical replication (WAL) | `pgoutput` |
| **MySQL** | Binlog (row-based) | MySQL connector |
| **SQL Server** | CDC capture job | SQL Server connector |
| **Oracle** | LogMiner / XStream | Oracle connector |
| **MongoDB** | Oplog | MongoDB connector |

### 1.3 Pipeline lengkap (visual detail)

```text
┌────────────────────────────────────────────────────────────────────────────┐
│                              LAYER 1: SOURCE                                │
│                                                                             │
│   ┌────────────┐   ┌────────────┐   ┌─────────────┐   ┌────────────┐       │
│   │ PostgreSQL │   │   MySQL    │   │ SQL Server  │   │   Oracle   │       │
│   │   WAL      │   │  binlog    │   │  CDC job    │   │  LogMiner  │       │
│   └─────┬──────┘   └─────┬──────┘   └──────┬──────┘   └──────┬─────┘       │
└─────────┼───────────────┼─────────────────┼─────────────────┼──────────────┘
          │               │                 │                 │
          │ Logical Rep   │ Replica         │ CDC             │ XStream
          │               │                 │                 │
          ▼               ▼                 ▼                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                    LAYER 2: CAPTURE (Debezium Connect)                     │
│                                                                             │
│   ┌──────────────────────────────────────────────────────────────────┐     │
│   │  Kafka Connect (JVM process)                                     │     │
│   │  ┌────────────────┐  ┌────────────────┐  ┌─────────────────┐    │     │
│   │  │ pg-source      │  │ mysql-source   │  │ mssql-source    │    │     │
│   │  │ task 1         │  │ task 1         │  │ task 1          │    │     │
│   │  └────────────────┘  └────────────────┘  └─────────────────┘    │     │
│   └──────────────────────────────────────────────────────────────────┘     │
│         │                 │                  │                              │
│         │ JSON event      │ JSON event       │ JSON event                  │
│         ▼                 ▼                  ▼                              │
└─────────┼─────────────────┼──────────────────┼──────────────────────────────┘
          │                 │                  │
          │                 │                  │
          ▼                 ▼                  ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                      LAYER 3: EVENT BUS (Kafka)                            │
│                                                                             │
│   Topics:                                                                   │
│   ┌──────────────────────────┐  ┌──────────────────────────┐                │
│   │ oltpdemo.public.orders   │  │ oltpdemo.shop.customers  │                │
│   │ oltpdemo.public.products │  │ oltpdemo.shop.orders     │                │
│   │ oltpdemo.dbo.Customers   │  │ oltpdemo.dbo.Orders      │                │
│   └──────────────────────────┘  └──────────────────────────┘                │
│                                                                             │
│   Schema Registry ◀──── JSON Schema (optional)                              │
│   Kafka Connect internal topics: connect-configs, connect-offsets           │
└──────────────────────────────────┬─────────────────────────────────────────┘
                                   │
                                   │ consume via Kafka Engine
                                   ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                       LAYER 4: ANALYTICAL (ClickHouse)                      │
│                                                                             │
│   Kafka Engine table: orders_kafka                                           │
│            │                                                                │
│            ▼                                                                │
│   Materialized View: mv_orders_raw → orders_raw (MergeTree)                 │
│            │                                                                │
│            ▼                                                                │
│   Query layer: orders_final (view with JSON extraction)                      │
│                                                                             │
│   Contoh event dari Debezium:                                                │
│   {                                                                         │
│     "before": null,                                                         │
│     "after": { "id": 123, "total": 45000, "status": "PAID" },               │
│     "op": "c",                                                              │
│     "ts_ms": 1717271234567                                                  │
│   }                                                                         │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Orkestrasi & Cara Konek

### 2.1 Single-server topology (development / demo)

Semua jalan di satu host:

```text
┌────────────────────────────────────────────────┐
│  Server 1 (8GB RAM)                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │  Kafka   │  │ Connect  │  │ ClickHse │     │
│  │  KRaft   │  │ Debezium │  │          │     │
│  └──────────┘  └──────────┘  └──────────┘     │
│  ┌──────────────┐  ┌──────────┐                │
│  │ Schema Reg.  │  │ Kafka UI │                │
│  └──────────────┘  └──────────┘                │
└────────────────────────────────────────────────┘
         ▲
         │ CDC
         │
   ┌─────┴──────┐
   │ RDS / DB   │
   └────────────┘
```

Cocok untuk:
- Demo / development
- Data volume < 1000 events/detik
- Budget minim

### 2.2 Multi-server topology (production scale-out)

Gak mungkin lempar semua di satu server kalau traffic gede. Pemisahan logis:

```text
┌────────────────────────────────────────────────────────────────────────────┐
│                            SOURCE LAYER (DB Servers)                        │
│                                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                       │
│  │ PG primary   │  │ MySQL primary│  │ MSSQL primary│                       │
│  │ + replica    │  │ + replica    │  │ + replica    │                       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                       │
└─────────┼─────────────────┼─────────────────┼───────────────────────────────┘
          │ replica slot    │ binlog dump     │ CDC capture
          ▼                 ▼                 ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                         INGESTION CLUSTER                                   │
│                                                                             │
│  Server A (capture)            Server B (stream)        Server C (storage)  │
│  ┌──────────────────┐         ┌──────────────────┐     ┌──────────────┐   │
│  │ Kafka Connect    │         │ Kafka broker     │     │ ClickHouse   │   │
│  │ Debezium         │────────▶│ 1 (controller+   │────▶│ shard 1      │   │
│  │ tasks: pg, mysql │         │  broker combined)│     │              │   │
│  │                  │         │                  │     │              │   │
│  │ 4GB RAM          │         │ 4GB RAM          │     │ 4GB RAM      │   │
│  └──────────────────┘         └──────────────────┘     └──────────────┘   │
│                                                                             │
│  Server D (consume)          Server E (HA)                                 │
│  ┌──────────────────┐         ┌──────────────────┐                        │
│  │ Kafka Connect    │         │ Kafka broker     │                        │
│  │ Debezium         │         │ 2                │                        │
│  │ tasks: mssql     │         │                  │                        │
│  │                  │         │ 4GB RAM          │                        │
│  │ 4GB RAM          │         └──────────────────┘                        │
│  └──────────────────┘                                                    │
└────────────────────────────────────────────────────────────────────────────┘
          │                              │
          ▼                              ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                       CLICKHOUSE CLUSTER (CH shard 2/3)                     │
│                                                                             │
│   Server C (shard 1)   Server F (shard 2)   Server G (shard 3)              │
│   ┌──────────────┐    ┌──────────────┐     ┌──────────────┐                  │
│   │ ClickHouse   │    │ ClickHouse   │     │ ClickHouse   │                  │
│   │ + ZK coord   │    │              │     │              │                  │
│   └──────────────┘    └──────────────┘     └──────────────┘                  │
│                                                                             │
│   ReplicatedMergeTree engine untuk HA                                      │
│   Distributed table di atas shard                                          │
└────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 Multi-server: "masa satu service yang sama?"

**TIDAK.** Setiap service jalan independen, scale terpisah. Konsepnya:

1. **Kafka Connect** bisa di-scale horizontal. Tambah worker baru = tambah paralel connector. 1 worker = 1 JVM process, bisa handle beberapa connector task sekaligus.
2. **Kafka broker** scale per-partisi. Topik dipecah jadi partisi, didistribusi ke broker. 1 broker = 1 process, 3 broker = 3 process.
3. **ClickHouse** scale per-shard. Tiap shard = 1 instance terpisah, query-nya pakai `Distributed` table engine.

Jadi kalo event/second naik:
- Tambah Kafka broker → kapasitas topic naik
- Tambah Kafka Connect worker → jumlah parallel connector task naik
- Tambah ClickHouse shard → query throughput naik

### 2.4 Koneksi antar service (network)

**Gak pakai `localhost` antar service kalau multi-server**, pakai hostname/container name:

| Dari | Ke | Protokol | Default port |
|---|---|---|---|
| Debezium | Kafka broker | PLAINTEXT/SASL_SSL | 9092 |
| Debezium | Source DB | native driver | 5432/3306/1433 |
| ClickHouse Kafka Engine | Kafka broker | PLAINTEXT/SASL_SSL | 9092 |
| App/client | ClickHouse HTTP | HTTP/HTTPS | 8123 |
| App/client | ClickHouse native | TCP | 9000 |
| Schema Registry | Kafka | PLAINTEXT | 9092 |
| Kafka UI | Kafka, Connect, SR | HTTP | 8080/8083/8081 |

Semua hostname di `docker-compose.yml` adalah **service name** (`kafka`, `connect`, `clickhouse`) yang otomatis resolve oleh Docker DNS. Di luar Docker, edit `/etc/hosts` atau pakai IP internal VPC.

### 2.5 Naming convention topic Kafka

```text
{topic_prefix}.{database}.{table}

Contoh:
- oltpdemo.public.orders
- oltpdemo.shop.customers
- oltpdemo.dbo.Orders
```

Prefix bisa diganti di `topic.prefix` connector config. Default: `oltpdemo` (biar gak bentrok kalau ada multiple pipeline).

### 2.6 Event format dari Debezium

Setiap perubahan row = 1 message JSON. Schema bisa dilihat di Kafka UI atau `kafka-console-consumer`:

```json
{
  "before": null,
  "after": {
    "id": 123,
    "customer_id": 45,
    "total": 45000,
    "status": "PAID",
    "created_at": 1717271234567
  },
  "source": {
    "version": "2.7.0.Final",
    "connector": "postgresql",
    "name": "oltpdemo",
    "ts_ms": 1717271234567,
    "snapshot": "false",
    "db": "appdb",
    "schema": "public",
    "table": "orders",
    "txId": 567,
    "lsn": 24567890
  },
  "op": "c",
  "ts_ms": 1717271234567,
  "transaction": null
}
```

`op`:
- `c` = create (insert)
- `u` = update
- `d` = delete
- `r` = read (initial snapshot)

`before` null untuk `c`, `after` null untuk `d`.

---

## 3. Cara Install (Step by Step)

### 3.1 Prasyarat

**Server target (bukan VPS DDAG 908MB):**

```bash
# Minimum
- 4 vCPU
- 8GB RAM
- 50GB SSD
- Ubuntu 22.04 LTS / Debian 12
- podman + podman-compose (atau docker + docker-compose)
- open port: 8123, 9000, 9092, 8083, 8081, 8088 (atau via reverse proxy)
```

**Source DB requirements:**

```sql
-- PostgreSQL (pg_hba.conf + postgresql.conf)
wal_level = logical
max_replication_slots = 4
max_wal_senders = 4

CREATE USER debezium WITH REPLICATION PASSWORD '...';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO debezium;

-- MySQL (my.cnf)
[mysqld]
server-id         = 1
log_bin           = mysql-bin
binlog_format     = ROW
binlog_row_image  = FULL
gtid_mode         = ON
enforce_gtid_consistency = ON

CREATE USER 'debezium'@'%' IDENTIFIED BY '...';
GRANT SELECT, RELOAD, SHOW DATABASES, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'debezium'@'%';

-- SQL Server
- Enable CDC at database level:  EXEC sys.sp_cdc_enable_db;
- Enable CDC at table level:    EXEC sys.sp_cdc_enable_table @source_schema='dbo', @source_name='Orders', @role_name=NULL;
```

### 3.2 Install stack

```bash
# 1. SSH ke server target
ssh user@your-server

# 2. Install podman (kalau belum)
sudo apt-get update
sudo apt-get install -y podman podman-compose

# 3. Clone repo
git clone https://github.com/RiprLutuk/ch-olap-pipeline.git
cd ch-olap-pipeline

# 4. Checkout branch
git checkout feature/kafka-debezium-architecture

# 5. Setup env
cd deploy/kafka-debezium
cp .env.example .env
nano .env   # edit credentials, host DB, dll

# 6. Bring up stack
podman compose up -d

# 7. Tunggu ~30 detik, lalu verify
./scripts/status.sh

# 8. Register Debezium connectors
./scripts/register-connectors.sh
```

### 3.3 Verifikasi

```bash
# Cek container jalan
podman ps

# Cek Kafka health
podman exec ch-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Cek connector status
curl -s http://127.0.0.1:8083/connectors | jq
curl -s http://127.0.0.1:8083/connectors/postgres-source/status | jq

# Cek ClickHouse
curl -s "http://127.0.0.1:8123/?query=SHOW+TABLES+FROM+analytics"

# Cek event masuk ke topic
podman exec ch-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic oltpdemo.public.orders \
  --from-beginning --max-messages 1

# Akses Kafka UI
# http://your-server:8088  (browse topics, schemas, connectors)
```

### 3.4 Buat tabel di ClickHouse otomatis

Saat container up, file di `clickhouse/*.sql` di-mount ke `/docker-entrypoint-initdb.d/` dan otomatis jalan pertama kali. Untuk nambah tabel/table baru:

```bash
# Via CLI
podman exec -it ch-clickhouse clickhouse-client

analytics=# CREATE TABLE orders_kafka (...);
```

Atau restart dengan file SQL baru di `clickhouse/`.

---

## 4. Troubleshooting

### 4.1 Connector gak start

```bash
# Lihat status detail
curl -s http://127.0.0.1:8083/connectors/postgres-source/status | jq

# Lihat error
curl -s http://127.0.0.1:8083/connectors/postgres-source/status | \
  jq '.tasks[0].trace'
```

**Common error: `connection refused` ke source DB**

Cek:
- Hostname resolve (`getent hosts your-db-host`)
- Port terbuka (`nc -zv your-db-host 5432`)
- Security group / firewall AWS allow inbound dari server ini
- SSL mode benar (`sslmode=require` untuk RDS)

### 4.2 Event gak masuk ke ClickHouse

```bash
# 1. Topic ada di Kafka?
podman exec ch-kafka kafka-topics --bootstrap-server localhost:9092 --list

# 2. Ada message di topic?
podman exec ch-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic oltpdemo.public.orders --from-beginning --max-messages 5

# 3. Kafka Engine table exist di CH?
curl -s "http://127.0.0.1:8123/?query=SHOW+TABLES+FROM+analytics"

# 4. Materialized view running?
curl -s "http://127.0.0.1:8123/?query=SELECT+count()+FROM+analytics.orders_raw"
```

**Kalau topic ada, ada message, tapi CH kosong:**

Cek `system.kafka_consumers`:
```sql
SELECT * FROM system.kafka_consumers;
```

Biasanya:
- `kafka_group_name` bentrok dengan instance CH lain → ganti group name
- `kafka_format` salah → pastikan `JSONAsString` atau `JSONEachRow`
- `kafka_broker_list` gak resolve dari dalam container CH → pakai service name, bukan IP

### 4.3 Source DB penuh dengan replication slot

Kalau pakai PostgreSQL dan consumer mati, WAL numpuk di `pg_replication_slots`:

```sql
-- Lihat slot
SELECT slot_name, active, restart_lsn FROM pg_replication_slots;

-- Hapus slot yang nyangkut (HATI-HATI, data bisa hilang)
SELECT pg_drop_replication_slot('debezium_pg_slot');
```

**Best practice**: monitor disk source DB. Kalau `wal_keep_size` default 16MB kehabisan, PostgreSQL bisa reject writes.

### 4.4 Memory habis

Tiap service JVM butuh memory tuning:

```yaml
# docker-compose.yml override
environment:
  KAFKA_HEAP_OPTS: "-Xms1g -Xmx2g"
  KAFKA_CONNECT_HEAP: "-Xms1g -Xmx2g"
  CLICKHOUSE_MAX_MEMORY_USAGE: 4000000000
```

Cek usage real-time:
```bash
podman stats
```

### 4.5 ClickHouse Kafka Engine stuck

```sql
-- Kill consumer yang nyangkut
KILL KAFKA CONSUMER WHERE table = 'orders_kafka';
DROP TABLE IF EXISTS orders_kafka;
-- Re-create
```

### 4.6 Reset full (kalau mau ulang dari nol)

```bash
./scripts/reset-demo.sh   # belum ada, manual:
podman compose down -v
rm -rf /tmp/kafka-data/* /tmp/clickhouse-data/*
podman compose up -d
```

### 4.7 Logging & observability

```bash
# Follow logs semua service
podman compose logs -f

# Specific service
podman logs -f ch-kafka
podman logs -f ch-connect
podman logs -f ch-clickhouse
```

Untuk production, tambahkan:
- Prometheus + Grafana untuk metric Kafka/CH
- Loki/ELK untuk log aggregation
- Alert: Kafka consumer lag > 1000, ClickHouse insert rate drop, source DB replication slot inactive

---

## 5. Production checklist

- [ ] Enable SASL_SSL untuk Kafka (bukan PLAINTEXT)
- [ ] Enable TLS untuk ClickHouse HTTPS
- [ ] Source DB user debezium = **read-only**, spesifik schema
- [ ] Set `wal_keep_size` Postgres ke 1GB+ (default 0 = unlimited, tapi safety)
- [ ] Set retention Kafka topic ke 7 hari (default gak ada)
- [ ] Monitor `kafka_consumergroup_lag`
- [ ] Backup ClickHouse pakai `BACKUP ... TO S3`
- [ ] Reverse proxy: jangan expose Kafka/Connect ke public, hanya via VPN
- [ ] Resource limits di compose: deploy.resources.limits.memory
- [ ] Healthcheck di compose untuk auto-restart
- [ ] Dashboard Grafana untuk: throughput, lag, error rate, DB source

---

## 6. Branching & deployment summary

| Branch | State | Action |
|---|---|---|
| `main` | Base + docs | stable |
| `archive/v1-cms-generator` | Native CMS+Gen snapshot | read-only |
| `feature/kafka-debezium-architecture` | Heavy stack work in progress | active dev |

**Next action items:**
1. Tambah `docker-compose.prod.yml` dengan SASL_SSL + TLS
2. Tambah `grafana/` dashboard JSON (Kafka lag, CH throughput)
3. Tambah `prometheus.yml` scrape config untuk Kafka exporter + CH
4. Tambah script backup CH ke S3
5. Tambah healthcheck di compose
6. Test end-to-end: insert ke RDS PG → muncul di CH dalam < 5 detik
