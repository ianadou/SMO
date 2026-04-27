# syntax=docker/dockerfile:1.7

# -----------------------------------------------------------------------------
# SMO server — multi-stage Dockerfile
#
# Stage 1 (builder): compiles a fully static Go binary with no C dependencies.
# Stage 2 (runtime): distroless static image running as a non-root user.
#
# The resulting image is ~20 MB and contains only the binary — no shell, no
# package manager, no libc. Minimal surface area for security.
# -----------------------------------------------------------------------------

# -----------------------------------------------------------------------------
# Stage 1 — builder
# -----------------------------------------------------------------------------
FROM golang:1.26.2-bookworm AS builder

WORKDIR /src

# Copy go.mod and go.sum first and download dependencies in a separate layer.
# This layer is cached as long as the module files do not change, which makes
# subsequent builds much faster when only source code is modified.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source tree.
COPY . .

# Build a fully static binary:
#   - CGO_ENABLED=0 → no libc dependency, required for distroless/static
#   - -trimpath     → strip local file system paths from the binary
#   - -ldflags      → strip DWARF debug info and symbol table for a smaller binary
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/smo-server \
    ./cmd/server

# Build the healthcheck probe as a separate static binary. Distroless has no
# shell, so the Dockerfile HEALTHCHECK directive below invokes this binary
# directly instead of relying on curl/wget.
RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/healthcheck \
    ./cmd/healthcheck

# -----------------------------------------------------------------------------
# Stage 2 — runtime
# -----------------------------------------------------------------------------
# The base image is pinned to an immutable SHA256 digest (not just the
# :nonroot tag) so that builds are reproducible across machines and time.
# To bump: `docker pull gcr.io/distroless/static-debian12:nonroot` and copy
# the new digest from the pull output.
FROM gcr.io/distroless/static-debian12@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1

# Metadata.
LABEL org.opencontainers.image.title="SMO server"
LABEL org.opencontainers.image.description="Sports Match Organizer HTTP server"
LABEL org.opencontainers.image.source="https://github.com/ianadou/SMO"
LABEL org.opencontainers.image.licenses="MIT"

# Copy the server and healthcheck binaries from the builder stage.
COPY --from=builder /out/smo-server /smo-server
COPY --from=builder /out/healthcheck /healthcheck

# Run as the non-root user provided by the distroless image (UID 65532).
USER nonroot:nonroot

# Document the default port (actual port is controlled by the PORT env var).
EXPOSE 8081

# Liveness probe. The healthcheck binary reads the PORT env var (default
# 8081) and GETs /health. Tuning rationale: a 10s interval keeps Dockhand
# responsive without flooding the app, the 3s timeout covers cold-start
# latency on a constrained VPS, start-period gives the server room to
# connect to Postgres and run migrations on boot.
HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=3 \
    CMD ["/healthcheck"]

# The binary is the entrypoint; no shell is available in distroless.
ENTRYPOINT ["/smo-server"]
