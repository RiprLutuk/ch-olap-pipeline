# Archive: v1 CMS + Generator (Native)

> **Status: ARCHIVED — tidak untuk deployment baru.**
> Pivot ke arsitektur **Kafka + Debezium** (lihat `deploy/kafka-debezium/`).

## Snapshot state terakhir (commit ini)

Repository berisi:

- `cmd/cms/` — Go HTTP CMS panel (`:8084`)
- `cmd/generator/` — Go generator multi-DB (`:8085`)
- `deploy/vps/install-native.sh` — installer native systemd (TANPA container)
- `deploy/vps/PEERDB_SETUP.md` — catatan PeerDB (alternatif CDC)
- `deploy/vps/CMS_BRIDGE.md` — catatan mode CH external vs self-hosted

## Yang pernah di-deploy (snapshot Juni 2026)

Server: VPS DDAG 908MB (`demo.ppn.pandanteknik.com`)

| Service | Port | Status terakhir |
|---|---|---|
| `ch-cms` | `:8084` | active, lalu **dropped** |
| `ch-gen` | `:8085` | active, lalu **dropped** |
| Caddy `/cms/` | public | lalu **restored** ke state sebelum route ini |

Generator berhasil konek ke RDS Postgres (`database-1.cwt8i4qk6wnt.us-east-1.rds.amazonaws.com`).
Caddy validate sukses.

## Alasan pivot

| Aspek | v1 (CMS+Gen native) | v2 (Kafka+Debezium) |
|---|---|---|
| RAM VPS 908MB | muat, tapi cuma sisa 60Mi free | TIDAK muat (Kafka JVM butuh 1.5GB+) |
| Real CDC support | partial (snapshot + periodic) | full (WAL/binlog streaming) |
| DB sources | PG, MySQL, SQL Server (koneksi langsung) | PG, MySQL, **SQL Server**, Oracle, MongoDB |
| Latency | 1-5 detik batch | <1 detik streaming |
| Target server | VPS 908MB kita | server dedicated 4-8GB RAM (di luar VPS ini) |

## Cara kerja v1 (ringkasan)

```
[RDS Postgres] ──direct──▶ [ch-gen :8085] ──batch──▶ [ClickHouse]
                              │
                              └─HTTP──▶ [ch-cms :8084] (UI monitoring)
```

Tanpa Kafka, tanpa Debezium, tanpa JVM. Cuma native Go binaries + systemd.
Cocok untuk demo kecil-kecilan & development, **TIDAK** untuk real CDC production.

## Migrasi ke v2

Lihat `deploy/kafka-debezium/README.md` di branch `main` untuk arsitektur baru.
