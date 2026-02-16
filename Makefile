.PHONY: all up down reset local test postgres app_logs postgres_logs lint .env .env.example help
.POSIX:
.SILENT:

-include .env.example .env

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
	docker compose up -d postgres app
	rm -f Dockerfile

down:
	docker compose down 2>/dev/null || true 
	rm -f Dockerfile docker-compose.yaml config.yaml

reset:
	docker volume rm proteus_minio-data

local:
	if [ ! -f .env ]; then cat .env.example > .env; fi 
	if [ ! -f config.yaml ]; then cp ./configs/config.dev.yaml ./config.yaml; fi 
	if [ ! -f docker-compose.yaml ]; then cp ./deployments/docker-compose.dev.yaml ./docker-compose.yaml; fi
	docker compose up -d
	until docker exec postgres pg_isready -U ${DB_USER} > /dev/null 2>&1; do sleep 0.5; done
	bash -c 'trap "exit 0" INT; go run ./cmd/proteus/main.go'

test:
	if [ ! -f .env ]; then cat .env.example > .env; fi 
	if [ ! -f config.yaml ]; then cp ./configs/config.test.yaml ./config.yaml; fi 
	if [ ! -f docker-compose.yaml ]; then cp ./deployments/docker-compose.test.yaml ./docker-compose.yaml; fi
	docker compose -f docker-compose.yaml up -d postgres-test
	until docker exec postgres-test pg_isready -U ${DB_USER} -d postgres-test > /dev/null 2>&1; do sleep 0.5; done
	echo "Running tests, please be patient (≈2 min)"
	docker compose -f docker-compose.yaml run --rm app-test > .temp 2>/dev/null
	cat .temp; rm -f .temp
	docker compose -f docker-compose.yaml down -v > /dev/null 2>&1
	rm -f docker-compose.yaml config.yaml .env

postgres:
	docker compose exec postgres psql -U ${DB_USER} -d proteus-db

app_logs:
	docker compose logs --tail 5 app

postgres_logs:
	docker compose logs --tail 5 postgres

lint:
	golangci-lint run ./...

.env:
	@:

help:
	@echo " ———————————————————————————————————————————————————————————————————————————————————— "
	@echo "| up             | Start all services (postgres, app) in background                  |"
	@echo "| down           | Stop and remove all containers, networks, and temporary files     |"
	@echo "| reset          | Remove postgres Docker volume                                     |"
	@echo "| local          | Start local dev environment (go 1.25.1 required)                  |"
	@echo "| test           | Run unit and integration tests                                    |"
	@echo "| postgres       | Open psql shell inside postgres container                         |"
	@echo "| app_logs       | Show last 5 lines of app logs                                     |"
	@echo "| postgres_logs  | Show last 5 lines of postgres logs                                |"
	@echo "| lint           | Run golangci-lint                                                 |"
	@echo " ———————————————————————————————————————————————————————————————————————————————————— "

.DEFAULT:
	@echo " No rule to make target '$@'. Available make targets:"
	@$(MAKE) --no-print-directory help