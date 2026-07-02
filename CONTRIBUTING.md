# Contributing to ch-olap-pipeline

First off, thank you for considering contributing. This project is meant to be a **collaborative, open-source platform for the DBA, data engineering, and platform engineering community** — your experience, your database, your connector, your case study is exactly what makes this useful for everyone else.

This document explains how to participate, what kinds of contributions are most valuable, and the workflow we follow.

---

## Table of contents

- [Code of Conduct](#code-of-conduct)
- [Project values](#project-values)
- [How can I contribute?](#how-can-i-contribute)
- [Adding a new source database adapter](#adding-a-new-source-database-adapter)
- [Adding a new sink adapter](#adding-a-new-sink-adapter)
- [Reporting bugs](#reporting-bugs)
- [Suggesting features](#suggesting-features)
- [Pull request workflow](#pull-request-workflow)
- [Commit message conventions](#commit-message-conventions)
- [Style guide](#style-guide)
- [Testing](#testing)
- [Documentation contributions](#documentation-contributions)
- [Community](#community)

---

## Code of Conduct

This project follows the [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md). By participating, you agree to uphold it. Be kind. Assume good faith. Disagree on ideas, not on people.

---

## Project values

1. **Operability first** — anything we ship must be deployable, observable, and recoverable.
2. **Modular adapters** — each source database is an isolated configuration + docs, not a fork of the core.
3. **Production docs** — we document failures, not just happy paths.
4. **Small, focused PRs** — easier to review, easier to revert, easier to learn from.

---

## How can I contribute?

There are many ways to help, and not all of them are code:

- **Add a new source database adapter** (Oracle, Db2, MariaDB, CockroachDB, etc.)
- **Add a new sink adapter** (BigQuery, Snowflake, Iceberg, Delta, etc.)
- **Improve or translate documentation**
- **Report bugs and edge cases** you hit in real environments
- **Submit deployment playbooks** for specific cloud providers
- **Share reproducible test data** for new database engines
- **Review pull requests**, especially from new contributors
- **Triage issues** by adding labels, reproduction steps, or related links
- **Improve observability** (Grafana dashboards, Prometheus exporters, alert rules)

If you are not sure where to start, look at issues tagged `good first issue` or `help wanted`.

---

## Adding a new source database adapter

This is the most impactful contribution. To add a new database, follow these steps:

### 1. Open an issue first

Before writing code, open an issue that explains:

- Which database (engine + version)
- What CDC mechanism it uses (WAL, binlog, native CDC, log-based, etc.)
- Whether a Debezium connector exists for it, or if a custom adapter is needed
- Any licensing, ops, or footprint caveats

This avoids duplicated work and gives the community a chance to weigh in.

### 2. Create a branch

```bash
git checkout -b feature/add-<db>-source-adapter
```

### 3. Add files under `deploy/kafka-debezium/connectors/`

Create `<db>-source.json` with the connector config template, using `envsubst` placeholders for credentials.

Example:

```json
{
  "name": "oracle-source",
  "config": {
    "connector.class": "io.debezium.connector.oracle.OracleConnector",
    "database.hostname": "${ORACLE_HOST}",
    "database.port": "${ORACLE_PORT}",
    "database.user": "${ORACLE_USER}",
    "database.password": "${ORACLE_PASSWORD}",
    "database.dbname": "${ORACLE_SID}",
    "topic.prefix": "oltpdemo",
    "table.include.list": "SCHEMA.TABLE",
    "log.mining.archive.log.hours": "24"
  }
}
```

### 4. Add an entry to `docs/supported-rdbms-matrix.md`

Move the database from `planned` to `beta` or `demo-ready` and link the connector config.

### 5. Add a section to `deploy/kafka-debezium/README.md`

Include:

- **Source DB requirements** (config flags, privileges, CDC enablement steps)
- **Connector config reference**
- **Known limitations**
- **Troubleshooting tips specific to this DB**

### 6. Submit a pull request

Follow the [pull request workflow](#pull-request-workflow) below.

---

## Adding a new sink adapter

Right now ClickHouse is the only sink. To add another:

1. Open an issue describing the sink, version, ingestion pattern (Kafka Engine / Kafka Connect Sink / JDBC / streaming insert).
2. Add files under `deploy/kafka-debezium/sinks/` or a parallel folder.
3. Document the schema mapping strategy, the dedup / out-of-order handling, and the verification steps.
4. Add a section to `deploy/kafka-debezium/README.md`.

---

## Reporting bugs

A good bug report saves everyone time. Please include:

- **Clear title** — one sentence describing the issue
- **Environment** — OS, podman / docker version, RAM, source DB version
- **Steps to reproduce** — exact commands, configs, and inputs
- **Expected behavior**
- **Actual behavior** — including full error messages, log lines, and `kafka-topics` / `clickhouse-client` output
- **Workarounds tried** (if any)

If the bug is about **data loss, gap, or schema mismatch**, mark the issue `priority: high` and include the **last known LSN / binlog position** and the **last event timestamp observed in ClickHouse**.

---

## Suggesting features

Open an issue with the `enhancement` label. Describe:

- The **problem** you are trying to solve, not just the solution
- Your **current workaround**, if any
- The **target users** who would benefit
- Any **prior art** (Debezium docs, blog posts, GitHub issues) that informed the idea

---

## Pull request workflow

1. **Fork** the repository.
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/<short-name>
   ```
3. **Make focused commits** — one logical change per commit. Use the [commit message conventions](#commit-message-conventions).
4. **Test locally** before pushing. At minimum:
   - `podman compose config` validates the compose file
   - `./scripts/status.sh` returns healthy state
   - `./scripts/register-connectors.sh` works with your new adapter
5. **Update documentation** — if you add a connector, add a section; if you change a config flag, update the README.
6. **Push** and **open a pull request** against `main`. Use the PR template.
7. **Respond to review feedback** — small PRs usually need 1-2 rounds. Larger ones may need more.

We aim to review PRs within 7 days. If you don't hear back, ping in the issue.

---

## Commit message conventions

We use **Conventional Commits** with a small twist:

```text
<type>(<scope>): <short summary>

<optional body explaining the why, not the what>

<optional footer with breaking change notes or issue refs>
```

Common types:

- `feat` — new feature (e.g. `feat(connectors): add oracle source adapter`)
- `fix` — bug fix
- `docs` — documentation only
- `refactor` — code change that neither fixes a bug nor adds a feature
- `chore` — tooling, CI, repo hygiene
- `test` — adding or fixing tests

Scopes we use:

- `connectors` — Debezium source / sink connector configs
- `clickhouse` — DDL, MVs, Kafka Engine tables
- `compose` — docker / podman compose files
- `scripts` — shell scripts
- `docs` — top-level docs under `docs/`
- `readme` — main README and visual assets

Keep the subject line **under 72 characters**. Use the body to explain motivation, not mechanics.

---

## Style guide

- **Shell scripts** — `set -euo pipefail`, no silent failures, prefer `envsubst` over hardcoded secrets.
- **JSON configs** — 2-space indent, alphabetical keys where possible, placeholders via `${ENV_VAR}`.
- **SQL** — uppercase keywords, lowercase identifiers, one statement per line.
- **Markdown** — wrap code blocks with a language tag, use tables for structured comparison, prefer real Markdown tables over ASCII art.

---

## Testing

Right now we are early-stage; automated tests are still being added. Until then, the minimum bar is:

- `podman compose config` passes
- `./scripts/status.sh` returns a healthy stack
- `./scripts/register-connectors.sh` registers without error
- A short end-to-end run is documented in the PR description (insert into source DB → verify event lands in ClickHouse within expected latency)

If you want to help build out the test harness, that is itself a great contribution. See the roadmap.

---

## Documentation contributions

Docs are first-class. You can contribute:

- **Typos, broken links, outdated commands**
- **Translated READMEs** — Bahasa Indonesia, Mandarin, Portuguese, Spanish, etc. are very welcome for the international community
- **Case studies** — “How we used this to migrate off a custom ETL pipeline at company X”
- **Cloud-specific deploy guides** — Hetzner, DigitalOcean, AWS, GCP, Azure
- **Screenshots and diagrams** — hand-drawn style works fine; nothing fancy required

---

## Community

- **Issues** for bug reports and feature requests
- **Discussions** for design debates, “does anyone have a recipe for X?” questions
- **Pull request reviews** for active contributors

There is no formal hierarchy. Long-term contributors gain triage / merge rights organically.

---

Thank you for reading this far. Even if you only fix a typo, you are helping someone save an afternoon. That matters.
