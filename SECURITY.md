# Security Policy

## Supported versions

| Branch | Supported |
|---|---|
| `main` | ✅ yes |
| `feature/kafka-debezium-architecture` | ✅ yes |
| `archive/v1-cms-generator` | ❌ no (read-only snapshot) |

## Reporting a vulnerability

If you discover a security issue in `ch-olap-pipeline`, please **do not open a public GitHub issue** for it. Security reports need a private channel so we can investigate and ship a fix before public disclosure.

Send your report to:

📧 **[INSERT_PROJECT_EMAIL_HERE]**

Please include:

- A clear description of the vulnerability
- Steps to reproduce (proof-of-concept preferred)
- The affected version / branch / commit SHA
- Your assessment of the impact (data exposure, privilege escalation, denial of service, etc.)
- Whether you intend to disclose publicly and on what timeline

We aim to acknowledge reports within **3 business days** and provide a triage assessment within **7 days**. Coordinated disclosure timelines are negotiable for serious issues.

## Security posture of this project

`ch-olap-pipeline` is a **data pipeline platform**, not a consumer-facing application. The primary security concerns are:

- **Source database credentials** — operators must use low-privilege CDC users; the project documents and requires this in every install guide
- **Network exposure of Kafka and Kafka Connect** — these services have no built-in auth in default config; the project binds them to localhost and documents opt-in TLS / SASL_SSL for remote access
- **ClickHouse exposure** — ClickHouse's HTTP and native protocols are unauthenticated in default config; the project documents the requirement to add users, restrict bind address, and use TLS
- **Secrets management** — `.env.example` shows the shape of secrets; actual `.env` files are git-ignored. Operators are expected to use a secret store (Vault, AWS Secrets Manager, etc.) in production
- **Debezium connector config** — credentials are passed via `${ENV_VAR}` substitution at registration time and never written to disk in plain text

The project **does not log credentials** at INFO level. Any debug-level log line that includes connection details is a bug and should be reported.

## Out of scope

- Vulnerabilities in **upstream dependencies** (Debezium, Kafka, ClickHouse, Podman, Docker). Please report those to the respective maintainers.
- Issues that require the operator to have already deployed with insecure defaults (e.g. exposing Kafka publicly without auth) — these are operator misconfigurations, not project bugs.

## Recognition

We maintain a list of security contributors who have helped improve the project. With your permission, we will credit you in release notes and the README once the report is resolved.
