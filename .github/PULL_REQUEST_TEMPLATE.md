## Summary

<!-- What does this PR change? -->

## Type of change

- [ ] Bug fix
- [ ] New source DB adapter
- [ ] New sink adapter
- [ ] Documentation
- [ ] CI / tooling
- [ ] Refactor
- [ ] Other

## Related issue

Closes #

## How was this tested?

Please include exact commands and outputs where possible.

```bash
# example
podman compose config
./scripts/status.sh
./scripts/register-connectors.sh
```

## CDC / data correctness checklist

If this PR touches ingestion, connectors, Kafka topics, or ClickHouse schemas:

- [ ] Primary key handling is documented
- [ ] Timestamp / offset handling is documented
- [ ] DELETE handling is documented
- [ ] Out-of-order / duplicate event behavior is documented
- [ ] Recovery path is documented
- [ ] Source DB privileges are documented

## Documentation checklist

- [ ] README updated if user-facing behavior changed
- [ ] `docs/supported-rdbms-matrix.md` updated if support status changed
- [ ] `deploy/kafka-debezium/README.md` updated if install/troubleshooting changed
- [ ] ADR added/updated if this changes architecture direction

## Security checklist

- [ ] No credentials or secrets are committed
- [ ] New network ports are not publicly exposed by default
- [ ] New DB user privileges are least-privilege where possible

## Screenshots / logs

<!-- Optional, but helpful for UI/docs/observability changes -->
