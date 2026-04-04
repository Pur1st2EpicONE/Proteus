.PHONY: all up down reset migrate-up migrate-down local lint test kafka minio postgres app_logs kafka_logs minio_logs postgres_logs .env .env.example help
.POSIX:
.SILENT:

-include .env.example .env

PROJECT_NAME = proteus
GOOSE_CMD = goose -dir ./migrations postgres "user=${DB_USER} password=${DB_PASSWORD} dbname=${PROJECT_NAME}-db host=localhost port=5433 sslmode=disable"

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
	COMPOSE_BAKE=true docker compose up -d
	docker exec ${PROJECT_NAME}-kafka-1 /opt/kafka/bin/kafka-topics.sh --create --if-not-exists --topic images --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
	rm -f Dockerfile

down:
	docker compose down 2>/dev/null || true 
	rm -f Dockerfile docker-compose.yaml config.yaml

reset:
	docker volume rm proteus_minio-data
	docker volume rm proteus_postgres_data

migrate-up:
	@if command -v goose > /dev/null 2>&1; then ${GOOSE_CMD} up; else echo "You need Goose migration tool to use this command!"; fi

migrate-down:
	@if command -v goose > /dev/null 2>&1; then ${GOOSE_CMD} down; else echo "You need Goose migration tool to use this command!"; fi

local:
	if [ ! -f .env ]; then cat .env.example > .env; fi 
	if [ ! -f config.yaml ]; then cp ./configs/config.dev.yaml ./config.yaml; fi 
	if [ ! -f docker-compose.yaml ]; then cp ./deployments/docker-compose.dev.yaml ./docker-compose.yaml; fi
	COMPOSE_BAKE=true docker compose up -d
	docker exec kafka /opt/kafka/bin/kafka-topics.sh --create --if-not-exists --topic images --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
	bash -c 'trap "exit 0" INT; go run ./cmd/${PROJECT_NAME}/main.go'

lint:
	golangci-lint run ./...

test:
	if [ ! -f .env ]; then cat .env.example > .env	; fi 
	if [ ! -f config.yaml ]; then cp ./configs/config.test.yaml ./config.yaml; fi 
	if [ ! -f docker-compose.yaml ]; then cp ./deployments/docker-compose.test.yaml ./docker-compose.yaml; fi
	COMPOSE_BAKE=true docker compose -f docker-compose.yaml up -d minio-test postgres-test
	until docker exec postgres-test pg_isready -U ${DB_USER} -d postgres-test > /dev/null 2>&1; do sleep 0.5; done
	echo "Running tests, please be patient (≈2 min)"
	COMPOSE_BAKE=true docker compose -f docker-compose.yaml run --rm app-test > .temp 2>/dev/null
	cat .temp; rm -f .temp
	docker compose -f docker-compose.yaml down -v > /dev/null 2>&1
	rm -f docker-compose.yaml config.yaml .env

kafka:
	docker compose exec kafka bash

minio:
	docker compose exec minio sh

postgres:
	docker compose exec postgres psql -U ${DB_USER} -d ${PROJECT_NAME}-db

app_logs:
	docker compose logs --tail 10 app

kafka_logs:
	docker compose logs --tail 10 kafka

minio_logs:
	docker compose logs --tail 10 minio

postgres_logs:
	docker compose logs --tail 10 postgres

.env:
	@:

help:
	@echo " ———————————————————————————————————————————————————————————————————————————————————— "
	@echo "| up             | Start all services (postgres, app, minio, kafka) in background    |"
	@echo "| down           | Stop and remove all containers, networks, and temporary files     |"
	@echo "| reset          | Remove postgres and minio Docker volumes                          |"
	@echo "| migrate-up     | Apply database migrations                                         |"
	@echo "| migrate-down   | Rollback last database migration                                  |"
	@echo "| local          | Start local dev environment (go 1.25.1 required)                  |"
	@echo "| lint           | Run golangci-lint                                                 |"
	@echo "| test           | Run unit and integration tests                                    |"
	@echo "| kafka          | Open bash shell inside kafka container                            |"
	@echo "| minio          | Open shell inside minio container                                 |"
	@echo "| postgres       | Open psql shell inside postgres container                         |"
	@echo "| app_logs       | Show last 10 lines of app logs                                    |"
	@echo "| kafka_logs     | Show last 10 lines of kafka logs                                  |"
	@echo "| minio_logs     | Show last 10 lines of minio logs                                  |"
	@echo "| postgres_logs  | Show last 10 lines of postgres logs                               |"
	@echo " ———————————————————————————————————————————————————————————————————————————————————— "

.DEFAULT:
	@echo " No rule to make target '$@'. Available make targets:"
	@$(MAKE) --no-print-directory help