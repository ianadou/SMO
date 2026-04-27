# ADR 0002 — Cache-aside with Redis

**Status:** Accepted (2026-04-27)

## Context

The two hottest read paths in SMO are `GroupRepository.FindByID` and
`PlayerRepository.ListByGroup`: every match-detail or group-detail
page hits both. The Postgres queries are cheap individually but
multiply with the frontend's polling pattern, and we want a buffer
ready before the deploy traffic arrives.

A Redis cache also lays the groundwork for the rate-limiting middleware
that lands in a follow-up PR — both features need a fast key-value
store with TTLs.

## Decision

### Pattern: cache-aside, decorator at the repository layer

Two new types — `CachingGroupRepository` and `CachingPlayerRepository`
— implement the existing domain ports and wrap the Postgres
repositories. The use cases see no change.

- **Read** (FindByID, ListByGroup): try Redis first; on miss or any
  Redis error, fall through to the inner repository and populate the
  cache with the fresh value.
- **Write** (Save, Update, UpdateRanking, Delete): delegate first;
  invalidate the affected key after the write succeeds. Order matters
  — DB first, cache second — so a cache deletion failure leaves the
  database authoritative.

Cached values are JSON-encoded `cachedGroup` / `cachedPlayer` structs
defined privately in the cache package. The domain entities never
gain JSON tags (CLAUDE.md architecture rule 1).

### TTL: 5 minutes

A safety net for cases where active invalidation is missed (e.g.,
direct database edits during ops). Active invalidation on every write
keeps the TTL mostly irrelevant in normal operation.

### Soft dependency at runtime

Redis errors at runtime are logged at WARN and the operation falls
through to Postgres. No 5xx is surfaced to the user. The cache is an
optimization, never a source of truth. `/health/ready` does not check
Redis: a Redis outage does not degrade the service contract.

### Invalidation on UpdateRanking

`CachingPlayerRepository.UpdateRanking` invalidates the affected
group's player list. This is non-negotiable: the FinalizeMatch use
case calls UpdateRanking once per participant, and without
invalidation the next ListByGroup would return stale rankings for up
to the TTL. Tested explicitly in
`player_repository_integration_test.go::TestCachingPlayerRepository_UpdateRanking_InvalidatesGroupListCache`.

### Cache key naming

- `smo:group:<groupID>` — single Group entity
- `smo:players:group:<groupID>` — list of Players in a group

The `smo:` prefix isolates SMO from any future co-tenant on the same
Redis instance.

## Configuration states

`REDIS_URL` controls three explicit states. Each is covered by a test.

| State | Trigger | Behavior | Test |
|-------|---------|----------|------|
| **1. Cache disabled** | `REDIS_URL` empty/unset | `Connect` returns `(nil, nil)`. The wiring helpers (`WrapGroupRepository`, `WrapPlayerRepository`) return the inner repo unchanged. No caching wrapper installed. Boot logs INFO `"redis cache disabled (REDIS_URL not set)"`. | `client_test.go::TestConnect_ReturnsNilClient_WhenURLIsEmpty` and `wiring_test.go::TestWrap*Repository_ReturnsInnerUnchanged_WhenClientIsNil` |
| **2. Boot failure** | `REDIS_URL` set, Redis unreachable or URL malformed | `Connect` returns `(nil, error)`. `run()` returns the error and the server refuses to boot. Aligns with CLAUDE.md axis 4 ("fail loud, fail with context"): a configured Redis that is unreachable is a misconfiguration, not graceful degradation. | `client_test.go::TestConnect_ReturnsError_WhenRedisIsUnreachable` and `TestConnect_ReturnsError_WhenURLIsMalformed` |
| **3. Runtime degradation** | Redis available at boot, becomes unavailable later | The caching repo logs WARN per failed Redis call and falls through to the inner Postgres repo. No error to the caller. | `group_repository_integration_test.go::TestCachingGroupRepository_FallsThroughToInner_WhenRedisGoesDown` |

## Consequences

- A new operational dependency (`redis:7-alpine` in `compose.yml`,
  later in the deploy stack). Health-checked.
- Cache invalidation bugs become a class of subtle staleness bugs: a
  forgotten invalidation surfaces only after 5 minutes. Mitigation:
  the wrappers are small and exhaustive, and integration tests verify
  every write path.
- Repository implementations become composable: future caches (e.g.,
  in-memory L1 in front of Redis) can be added by stacking decorators
  without touching the use cases.

## Alternatives considered and rejected

- **Cache injected into the use case via a `Cache` port** — couples
  the use case to caching. Tests of the use case need a fake cache.
  Decorator pattern at the repo layer keeps caching purely
  infrastructural.
- **Cache as a Gin middleware** — too coarse: the middleware cannot
  see business-level invalidation events (a Save on Group X must
  invalidate the read of Group X, but a generic HTTP cache cannot
  know that).
- **Hard dependency** (Redis down → 503) — rejected. The cache is an
  optimization. A best-effort fallback to Postgres preserves
  availability at the cost of latency, which is the right trade-off
  for SMO's traffic profile.
- **Per-key TTL** (e.g., 1 min for hot lists, 1 hour for stable
  groups) — premature complexity. A single 5-minute TTL with active
  invalidation is sufficient until profiling shows otherwise.
