.PHONY: up down rebuild register status logs ch-psql ch-cli mysql-cli \
        generator-logs generator-host build fmt test

COMPOSE ?= podman compose

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

rebuild:
	$(COMPOSE) up -d --no-deps --force-recreate --build generator

register:
	@echo "Registering Debezium connectors..."
	@for f in deploy/debezium/*.json; do \
	  name=$$(basename $$f .json); \
	  echo " → $$name"; \
	  curl -fsS -X POST -H "Content-Type: application/json" \
	    --data @$$f \
	    http://localhost:8083/connectors || true; \
	done

status:
	@curl -fsS http://localhost:8083/connectors | jq -r '.[] | "\(.name): \(.status.state // "UNKNOWN")"'

logs:
	$(COMPOSE) logs -f --tail=100

generator-logs:
	$(COMPOSE) logs -f --tail=100 generator

ch-cli:
	@docker exec -it olap-clickhouse clickhouse-client \
	  --user analytics --password analytics --database shop_analytics

mysql-cli:
	@docker exec -it oltp-mysql mysql -ushop -pshop shop

build:
	cd cmd/generator && go build -o ../../bin/generator .

fmt:
	gofmt -s -w .

test:
	go test ./...
