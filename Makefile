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
