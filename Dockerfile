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

# -----------------------------------------------------------------------------
# Stage 2 — runtime
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

# Metadata.
LABEL org.opencontainers.image.title="SMO server"
LABEL org.opencontainers.image.description="Sports Match Organizer HTTP server"
LABEL org.opencontainers.image.source="https://github.com/ianadou/SMO"
LABEL org.opencontainers.image.licenses="MIT"

# Copy the binary from the builder stage.
COPY --from=builder /out/smo-server /smo-server

# Run as the non-root user provided by the distroless image (UID 65532).
USER nonroot:nonroot

# Document the default port (actual port is controlled by the PORT env var).
EXPOSE 8081

# The binary is the entrypoint; no shell is available in distroless.
ENTRYPOINT ["/smo-server"]
