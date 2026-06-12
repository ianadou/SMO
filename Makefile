# SMO — Sports Match Organizer
# Common development tasks. Run `make help` for a list.

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the server binary
	go build -o /tmp/smo-server ./cmd/server

.PHONY: run
run: ## Run the server locally on port 8081
	go run ./cmd/server

.PHONY: lint
lint: ## Run golangci-lint on the whole codebase
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with --fix (auto-fixes what it can)
	golangci-lint run --fix ./...

.PHONY: fmt
fmt: ## Format the code with gofumpt
	gofumpt -w .

.PHONY: test
test: ## Run all unit tests
	go test ./...

.PHONY: test-integration
test-integration: ## Run unit + integration tests (requires Docker for testcontainers)
	go test -tags=integration ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -f /tmp/smo-server coverage.out

# ----------------------------------------------------------------------------
# Database migrations (goose)
# ----------------------------------------------------------------------------
# These targets require goose installed locally:
#   go install github.com/pressly/goose/v3/cmd/goose@latest
#
# DATABASE_URL must point to a running Postgres instance. The default value
# below assumes a local development database; override it for other envs:
#   make migrate-up DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable
DATABASE_URL ?= postgres://smo:smo@localhost:5433/smo_dev?sslmode=disable
MIGRATIONS_DIR := infrastructure/persistence/migrations

.PHONY: migrate-up
migrate-up: ## Apply all pending database migrations
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back the most recent migration
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

.PHONY: migrate-status
migrate-status: ## Show the status of all migrations
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

.PHONY: migrate-create
migrate-create: ## Create a new SQL migration file (use NAME=description)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=description_here"; exit 1; fi
	goose -dir $(MIGRATIONS_DIR) -s create $(NAME) sql

# ----------------------------------------------------------------------------
# Docker Compose
# ----------------------------------------------------------------------------
# Convenience targets around `docker compose`. They assume Docker Compose v2
# (the `docker compose` subcommand, not the legacy `docker-compose` binary).
#
# The compose stack defines two services:
#   - postgres : PostgreSQL 16 database with a named persistent volume
#   - app      : the SMO HTTP server, built from the local Dockerfile
#
# Environment variables come from .env at the repository root. `make env`
# creates it on first run with locally generated secrets, so a fresh
# clone boots with `make up` and nothing sensitive is ever committed.

.PHONY: env
env: ## Create .env from .env.example with generated local secrets (no-op if present)
	@test -f .env || { \
		cp .env.example .env; \
		sed -i "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$$(openssl rand -hex 16)|" .env; \
		sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$$(openssl rand -hex 32)|" .env; \
		echo ".env created with generated local secrets"; \
	}

.PHONY: up
up: env ## Clone-and-run entrypoint: build and start the full stack
	docker compose up -d --build
	@echo ""
	@echo "SMO is starting:"
	@echo "  Frontend  http://localhost:3000"
	@echo "  API       http://localhost:8081/api/v1"
	@echo "  Health    http://localhost:8081/health/ready"

.PHONY: compose-up
compose-up: env ## Start the full Docker Compose stack in the background
	docker compose up -d

.PHONY: compose-up-db
compose-up-db: ## Start only the postgres service (useful for `go run` against it)
	docker compose up -d postgres

.PHONY: compose-down
compose-down: ## Stop and remove containers (keeps the data volume)
	docker compose down

.PHONY: compose-reset
compose-reset: ## Stop containers AND delete the data volume (full reset)
	docker compose down -v

.PHONY: compose-logs
compose-logs: ## Tail the logs of all services
	docker compose logs -f

.PHONY: compose-logs-app
compose-logs-app: ## Tail only the app service logs
	docker compose logs -f app

.PHONY: compose-ps
compose-ps: ## Show the status of compose services
	docker compose ps
