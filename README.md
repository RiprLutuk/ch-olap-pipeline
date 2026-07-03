# CH OLAP Pipeline

Beginner-friendly, community-first OLTP to OLAP pipeline for DBA, infra, and data-platform engineers.

`ch-olap-pipeline` is an open-source reference project for moving database changes from operational systems into analytical storage using real CDC patterns.

## What this project is

This repository is for people who want to learn, evaluate, and gradually build a practical pipeline like:

`PostgreSQL / MySQL / SQL Server / APIs -> Debezium / Kafka Connect -> Kafka -> ClickHouse`

It is designed to be:

- **human friendly** — clear docs, simple concepts, honest scope
- **beginner friendly** — obvious starting points, not just deep internals
- **ops friendly** — secure defaults, troubleshooting, production-minded notes
- **community extensible** — adapters and sink docs can grow over time

## Who should use this

Good fit for:

- DBAs learning CDC and analytical replication
- infra / backend engineers who want a practical reference
- data-platform teams evaluating Kafka + Debezium + ClickHouse
- teams moving away from cron polling ETL

Probably **not** the best fit if you want:

- a one-click SaaS product
- a fully finished UI-driven data integration platform
- full production support for every database listed in the docs today

## What works today vs what is scaffolded

### Works well as a reference today

- architecture direction and decision records
- source and sink matrices
- connector template layout
- ClickHouse-first design direction
- OSS governance and contributor docs
- docs build with MkDocs

### Still scaffold / roadmap heavy

- many non-core adapters are templates, not production-hardened implementations
- some connector examples are meant as starting points for contributors
- full end-to-end runtime validation is not bundled for every database/sink combination

So: this repo is already useful, but it is still closer to a **serious reference platform** than a finished plug-and-play product.

## Start here

If you are new, read in this order:

1. [`docs/index.md`](docs/index.md) — quick project overview
2. [`docs/system-overview.md`](docs/system-overview.md) — simple architecture explanation
3. [`docs/architecture-decisions.md`](docs/architecture-decisions.md) — why the design looks like this
4. [`docs/supported-rdbms-matrix.md`](docs/supported-rdbms-matrix.md) — current source coverage
5. [`docs/sinks-matrix.md`](docs/sinks-matrix.md) — current sink coverage

## Recommended first learning path

If you only want one clean learning path, start with:

- **Source:** PostgreSQL, MySQL, or SQL Server
- **Transport:** Debezium + Kafka Connect + Kafka
- **Sink:** ClickHouse

That is the main opinionated path for this repository.

## Deployment reality

This project is intentionally honest about infrastructure tradeoffs.

### Tiny VPS (like ~1GB RAM)

A full stack with Kafka + Connect + Debezium + ClickHouse is usually **not** a good idea on a very small VPS.

Tiny hosts are better for:

- docs and validation work
- light control-plane services
- small demos with externalized heavy components

### Better lab target

For realistic end-to-end testing, use:

- local machine / laptop
- a dedicated VM
- a 4–8GB lab host

That gives enough room for Kafka, Connect, and ClickHouse without constantly fighting RAM pressure.

## Security defaults

The architecture assumes secure-by-default behavior:

- Kafka / Connect / ClickHouse should stay localhost-only unless explicitly exposed
- source databases should use low-privilege CDC users
- public exposure should be intentional and authenticated

See [`SECURITY.md`](SECURITY.md) and the architecture docs for details.

## Repository contents

- `docs/` — architecture, vision, matrices, adapter docs, sink docs
- `deploy/` — example connector and deployment templates
- `.github/` — issue templates, discussions guide, PR template

## Development

Build docs locally:

```bash
python3 -m mkdocs build --clean --strict
```

GitHub Actions are intentionally stored disabled under `.github/workflows.disabled/` for this initial clean release.

## Contact and community

This project uses GitHub-native contact paths instead of fake public placeholder emails.

- Maintainer: https://github.com/RiprLutuk
- Discussions: https://github.com/RiprLutuk/ch-olap-pipeline/discussions
- Security advisories: https://github.com/RiprLutuk/ch-olap-pipeline/security/advisories

## License

MIT
