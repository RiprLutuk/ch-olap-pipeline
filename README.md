# CH OLAP Pipeline

Community-first OLTP to OLAP pipeline for DBA, infra, and data-platform engineers.

The target architecture is simple and practical:

`RDBMS / API sources -> Kafka Connect / Debezium -> Kafka -> ClickHouse`

The repository is intentionally documentation-first and scaffold-first so contributors can add adapters and deployment recipes cleanly.

## What is included

- Multi-RDBMS source matrix
- Sink target matrix
- Kafka/Debezium connector templates
- ClickHouse-first OLAP direction
- OSS governance files: contributing guide, code of conduct, security policy, contact guide, issue templates, PR template
- GitHub Actions templates stored disabled under `.github/workflows.disabled/`

## Source coverage

See [`docs/supported-rdbms-matrix.md`](docs/supported-rdbms-matrix.md).

## Sink coverage

See [`docs/sinks-matrix.md`](docs/sinks-matrix.md).

## Security and contact

This project does not publish fake placeholder emails. Use GitHub-native channels:

- Maintainer: https://github.com/RiprLutuk
- Discussions: https://github.com/RiprLutuk/ch-olap-pipeline/discussions
- Security advisories: https://github.com/RiprLutuk/ch-olap-pipeline/security/advisories

## Development

Build docs locally:

```bash
python3 -m mkdocs build --clean --strict
```

GitHub Actions are intentionally disabled for the initial clean release. Enable files from `.github/workflows.disabled/` only when the repository is ready for hosted CI/Pages.
