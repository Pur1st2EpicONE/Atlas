.PHONY: all up down reset local migrate-up migrate-down test postgres rabbit app_logs postgres_logs rabbit_logs_logs queues lint .env .env.example help
.POSIX:
.SILENT:

-include .env.example .env

GOOSE_CMD = goose -dir ./migrations postgres "user=${DB_USER} password=${DB_PASSWORD} dbname=atlas-db host=localhost port=5433 sslmode=disable"

all: up

up:	
	if [ ! -f .env ] && [ ! -f .env.example ]; then \
		echo "Missing environment file: .env or .env.example is required."; \
		exit 1; \
	fi
	if [ ! -f .env ]; then cat .env.example > .env; fi
	if [ ! -f config.yaml ]; then cp ./configs/config.full.yaml ./config.yaml; fi
	if [ ! -f docker-compose.yaml ]; then cp ./deployments/docker-compose.full.yaml ./docker-compose.yaml; fi
	if [ ! -f Dockerfile ]; then cp ./deployments/Dockerfile ./Dockerfile; fi
	COMPOSE_BAKE=true docker compose up -d postgres rabbitmq app
	rm -f Dockerfile

down:
	docker compose down 2>/dev/null || true 
	rm -f Dockerfile docker-compose.yaml config.yaml

reset:
	docker volume rm kairos_postgres_data

local:
	if [ ! -f .env ]; then cat .env.example > .env; fi 
	if [ ! -f config.yaml ]; then cp ./configs/config.dev.yaml ./config.yaml; fi 
	if [ ! -f docker-compose.yaml ]; then cp ./deployments/docker-compose.dev.yaml ./docker-compose.yaml; fi
	docker compose up -d postgres rabbitmq
	until docker exec rabbitmq rabbitmqctl status > /dev/null 2>&1; do sleep 0.5; done
	bash -c 'trap "exit 0" INT; go run ./cmd/atlas/main.go'

migrate-up:
	@if command -v goose > /dev/null 2>&1; then $(GOOSE_CMD) up; else echo "You need Goose migration tool to use this command!"; fi

migrate-down:
	@if command -v goose > /dev/null 2>&1; then $(GOOSE_CMD) down; else echo "You need Goose migration tool to use this command!"; fi

test:
	cat .env.example > .env
	cp ./configs/config.dev.yaml ./config.yaml
	cp ./deployments/docker-compose.dev.yaml ./docker-compose.yaml
	go test -cover ./internal/handler/v1/...
	go test -cover ./internal/service/impl/...
	docker compose -f docker-compose.yaml up -d postgres-test > /dev/null 2>&1
	until docker exec postgres-test pg_isready -U ${DB_USER} > /dev/null 2>&1; do sleep 0.5; done
	for i in $$(seq 1 10); do \
		migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@localhost:5434/kairos_test?sslmode=disable" up > /dev/null 2>&1 && exit 0; sleep 1; \
	done; exit 1
	go test ./internal/repository/postgres -cover
	docker compose -f docker-compose.yaml stop postgres-test > /dev/null 2>&1
	docker compose -f docker-compose.yaml rm -f postgres-test > /dev/null 2>&1
	rm -f docker-compose.yaml config.yaml .env

postgres:
	docker compose exec postgres psql -U ${DB_USER} -d atlas-db

rabbit:
	docker compose exec rabbitmq bash

app_logs:
	docker compose logs --tail 10 app

postgres_logs:
	docker compose logs --tail 10 postgres

rabbit_logs:
	docker compose logs --tail 10 rabbitmq

queues:
	docker compose exec rabbitmq rabbitmqctl list_queues

lint:
	golangci-lint run ./...

.env:
	@:

help:
	@echo " ———————————————————————————————————————————————————————————————————————————————————— "
	@echo "| up             | Start all services (postgres, rabbitmq, app) in background        |"
	@echo "| down           | Stop and remove all containers, networks, and temporary files     |"
	@echo "| reset          | Remove postgres Docker volume                                     |"
	@echo "| local          | Start local dev environment (go 1.25.1 required)                  |"
	@echo "| migrate-up     | Apply all database migrations (Goose migration tool required)     |"
	@echo "| migrate-down   | Rollback all database migrations (Goose migration tool required)  |"
	@echo "| test           | Run unit and integration tests                                    |"
	@echo "| postgres       | Open psql shell inside postgres container                         |"
	@echo "| rabbit         | Open shell inside rabbitmq container                              |"
	@echo "| app_logs       | Show last 10 lines of app logs                                    |"
	@echo "| postgres_logs  | Show last 10 lines of postgres logs                               |"
	@echo "| rabbit_logs    | Show last 10 lines of rabbitmq logs                               |"
	@echo "| queues         | List queues in rabbitmq                                           |"
	@echo "| lint           | Run golangci-lint                                                 |"
	@echo " ———————————————————————————————————————————————————————————————————————————————————— "

.DEFAULT:
	@echo " No rule to make target '$@'. Available make targets:"
	@$(MAKE) --no-print-directory help