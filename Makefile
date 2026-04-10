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
test: ## Run all tests
	go test ./...

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
DATABASE_URL ?= postgres://smo:smo@localhost:5432/smo_dev?sslmode=disable
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
