# SMO — Sports Match Organizer

A platform to organize sports matches, assign teams dynamically based on player ranking, and manage post-match peer voting.

## Status

🚧 Work in progress — built as a CI/CD & DevOps school project, maintained as a long-lived portfolio project.

## Stack

- **Backend:** Go 1.26, Gin, sqlc, PostgreSQL, Redis
- **Frontend:** Nuxt 3, TypeScript, Tailwind CSS
- **Infrastructure:** Docker, Terraform, Ansible, GitHub Actions, Prometheus, Grafana

## Architecture

Hexagonal (Ports & Adapters) — pure domain with explicit mappers to persistence models.

## Development setup

### Prerequisites

- Go 1.26 or later
- [golangci-lint](https://golangci-lint.run/) v2.11+
- [gofumpt](https://github.com/mvdan/gofumpt)
- [pre-commit](https://pre-commit.com/) (any recent version)

### Clone and bootstrap

\`\`\`bash
git clone git@github.com:ianadou/SMO.git
cd SMO
pre-commit install --install-hooks
\`\`\`

The \`pre-commit install\` step is **required**: it sets up local Git hooks that automatically run formatting, linting, and commit message validation before every commit. Skipping this step means your commits will not be checked locally and will likely fail in CI.

### Common commands

All daily development tasks are exposed via the \`Makefile\`:

\`\`\`bash
make help        # list all available targets
make build       # build the server binary
make run         # run the server locally on port 8081
make lint        # run golangci-lint on the whole codebase
make fmt         # format Go code with gofumpt
make test        # run the test suite
\`\`\`

### Running the server

\`\`\`bash
make run
# or with a custom port:
PORT=9090 make run
\`\`\`

The server exposes a \`/health\` endpoint:

\`\`\`bash
curl http://localhost:8081/health
# {"status":"ok"}
\`\`\`

## Documentation

Full documentation, architecture decisions, and roadmap will be published here progressively during development.

## License

MIT — see [LICENSE](./LICENSE).
