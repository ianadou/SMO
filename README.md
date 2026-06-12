# SMO — Sports Match Organizer

[![CI](https://github.com/ianadou/SMO/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/ianadou/SMO/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ianadou/SMO/branch/main/graph/badge.svg)](https://codecov.io/gh/ianadou/SMO)

A platform to organize sports matches, assign teams dynamically based on player ranking, and manage post-match peer voting.

## Status

🚧 Work in progress — built as a CI/CD & DevOps school project, maintained as a long-lived portfolio project.

## Stack

- **Backend:** Go 1.26, Gin, sqlc, PostgreSQL, Redis
- **Frontend:** Nuxt 3, TypeScript, Tailwind CSS
- **Infrastructure:** Docker, Terraform, Ansible, GitHub Actions, Prometheus, Grafana

## Architecture

Hexagonal (Ports & Adapters) — pure domain with explicit mappers to persistence models.

## Quick start (clone and run)

The only prerequisites are Docker (Compose v2) and `make`:

```bash
git clone git@github.com:ianadou/SMO.git
cd SMO
make up
```

`make up` generates a local `.env` with random secrets on first run
(nothing sensitive is committed), builds the images and starts the full
stack. Once healthy:

- Frontend: <http://localhost:3000>
- API: <http://localhost:8081/api/v1>
- Health: <http://localhost:8081/health/ready>

## Development setup

### Prerequisites

- Go 1.26 or later
- [golangci-lint](https://golangci-lint.run/) v2.11+
- [gofumpt](https://github.com/mvdan/gofumpt)
- [pre-commit](https://pre-commit.com/) (any recent version)

### Clone and bootstrap

```bash
git clone git@github.com:ianadou/SMO.git
cd SMO
pre-commit install --install-hooks
```

The `pre-commit install` step is **required**: it sets up local Git hooks that automatically run formatting, linting, and commit message validation before every commit. Skipping this step means your commits will not be checked locally and will likely fail in CI.

### Common commands

All daily development tasks are exposed via the `Makefile`:

```bash
make help              # list all available targets
make build             # build the server binary
make run               # run the server locally on port 8081
make lint              # run golangci-lint on the whole codebase
make fmt               # format Go code with gofumpt
make test              # run the unit test suite
make test-integration  # unit + integration tests (requires Docker)
```

### Running the server

```bash
make run
# or with a custom port:
PORT=9090 make run
```

The server exposes two health endpoints — `/health/live` (process up)
and `/health/ready` (dependencies checked):

```bash
curl http://localhost:8081/health/ready
# {"status":"ok","database":"ok","cache":"ok","timestamp":"..."}
```

## Documentation

- [Architecture](docs/architecture.md), [Design patterns](docs/design-patterns.md),
  [CI/CD pipeline](docs/ci-cd-pipeline.md), [Infrastructure](docs/infrastructure.md),
  [Monitoring](docs/monitoring.md), [Constraints & workarounds](docs/constraints-and-workarounds.md)
  (in French — module deliverable)
- [Architecture Decision Records](docs/adr/)
- [Mobile UI walkthrough](docs/screenshots/) — the full user story in screenshots

## License

MIT — see [LICENSE](./LICENSE).
