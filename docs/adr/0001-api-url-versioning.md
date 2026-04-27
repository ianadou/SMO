# ADR 0001 — API URL versioning

**Status:** Accepted (2026-04-27)

## Context

The HTTP API is the public contract between the SMO backend and its
clients (Nuxt frontend today, possibly a mobile app or third-party
integrations later). Once a client is deployed against a given URL
shape, breaking changes become expensive: every consumer must redeploy
in lockstep with the backend, or the change must be rolled out as a
gradual migration with feature flags.

Two mainstream versioning strategies were considered:

1. **URL-based** — the version is part of the path (e.g. `/api/v1/matches`).
2. **Header-based** — the version is a request header
   (e.g. `Accept: application/vnd.smo.v1+json`).

## Decision

We adopt **URL-based versioning**, with the `/v1` prefix applied to all
business endpoints under `/api/v1/...`. The `/health` endpoint stays at
the root because it is an infrastructure check, not part of the
business contract.

## Rationale

- **Discoverability**: a developer reading the routes list immediately
  sees the version. `curl /api/v1/groups` is self-documenting.
- **Tooling support**: OpenAPI generators, API gateways, reverse
  proxies, and request loggers all handle URL-based versioning out of
  the box.
- **Cost of adoption**: a single line change in the router. Adding it
  later, after clients are deployed, would require updating the
  frontend, the OpenAPI spec, every doc snippet, and the deploy
  scripts.
- **Project shape**: SMO has one client today (the Nuxt frontend) and
  no public API ambition. URL versioning ships ceremony commensurate
  with the audience; header versioning would be over-engineered for
  this scale.

## Consequences

- Every new business endpoint must live under `/api/v1/...`. Reviewers
  reject PRs that add routes outside this prefix.
- A future breaking change ships as `/api/v2/...` alongside `/v1`.
  `/v1` is then deprecated with a notice period (`Deprecation` and
  `Sunset` HTTP headers per RFC 8594).
- The `/health` endpoint is intentionally not versioned. It belongs to
  the infrastructure layer (used by Docker HEALTHCHECK, Kubernetes
  probes, Dockhand) and its shape is owned by ops, not product.
