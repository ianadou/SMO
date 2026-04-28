# ADR 0004 — Domain Events as the default pattern for cross-cutting reactions

**Status:** Accepted (2026-04-28)

## Context

Several cross-cutting concerns need to react to business state
transitions without leaking infrastructure into the domain or
multiplying constructor parameters of the use cases:

- Discord notification when a match transitions to `teams_ready` (PR #55)
- Prometheus counters and histograms on every state transition
- Account-level lockout audit on repeated failed logins (issue #49)
- Generic audit trail of significant business actions

A naive approach — passing a `Notifier` port to the use case, then a
`MetricsRecorder` port, then an `AuditLogger` port — couples the use
case to every concern and forces a constructor change every time a new
reaction is added. It also tempts callers to add infrastructure types
to the domain layer, which violates CLAUDE.md architecture rule 1.

## Decision

**Domain Events become the default pattern for any reaction to a
business state transition.** The use case publishes a value-typed event
through an `EventPublisher` port; subscribers in the infrastructure
layer register at boot time in `cmd/server/main.go` and react.

### Default for future cross-cutting concerns

Every future addition that observes business transitions must use this
pattern unless explicitly justified otherwise:

- **Prometheus metrics** — a `MetricsSubscriber` listening on every
  event name, no use case change required to add a counter
- **Account lockout** (#49) — a `LockoutSubscriber` on a future
  `LoginFailed` event
- **Audit log** — already shipped as `LoggingSubscriber` in this PR
- **Future Discord/Slack/email/webhook transports** — additional
  subscribers on the relevant event, no fan-out logic in the domain

The rule: **if the reaction is observable from the domain's emission
of an event, prefer a subscriber over a port injected into the use
case.** Direct ports remain valid only for things the domain
**actively requests** (a token, a clock, a hash, a row).

### In-process synchronous publisher

The `inmemory.Publisher` adapter is the only implementation. It
dispatches synchronously: `Publish` returns once every subscriber has
been invoked. No goroutines, no queue, no retry. This is appropriate
for a single-process Go app on Hetzner CAX11; introducing an async or
out-of-process variant (Kafka, NATS, Redis Streams) would warrant a
dedicated ADR.

A subscriber that returns an error is logged at WARN and the dispatch
continues with the next subscriber. A subscriber must be cheap and
non-blocking: a slow subscriber blocks the calling use case.

### Dispatch by EventName, not by Go type

The publisher dispatches based on `event.EventName() string`, not on
the Go type of the event. Subscribers register with `Subscribe(name,
sub)`.

| Aspect | Dispatch by name (chosen) | Dispatch by type (rejected) |
|---|---|---|
| Subscriber registration | `Subscribe("match.teams_ready", sub)` | reflect-based or generics-based |
| Observability | event name appears in logs/traces verbatim | type name leaks Go internals |
| Cross-language readiness | name is wire-stable if we ever export events | type-based dispatch is Go-only |
| Testability | one fake event with a custom `EventName()` is enough to test routing | every test needs a new Go type |
| Cost | a renamed event is a silent breaking change | a renamed type is a compile error |

The trade-off is explicit: **renaming an event constant
(`MatchTeamsReadyEventName`) is a breaking change that the Go compiler
does NOT detect.** Subscribers registered under the old name silently
stop receiving events, and the failure surfaces only in production.

**Mitigation:** any renaming of an event name must ship in a dedicated
PR with: (1) update of the `*EventName` constant in `domain/events/`,
(2) update of every `Subscribe(...)` call in `cmd/server/main.go`,
(3) an integration test asserting the subscriber is invoked end-to-end
under the new name.

### Logging strategy

`LoggingSubscriber` takes an injected `*slog.Logger` rather than
calling `slog.Default()` internally:

- **Testability** — tests pass a logger backed by a `bytes.Buffer` to
  capture output without monkey-patching the global logger
- **Consistency** — every other infrastructure component in the
  codebase (HTTP middlewares, repositories, rate limiter) takes its
  logger by injection
- **Request correlation** — subscribers receive the calling use
  case's `context.Context` via `Handle(ctx, event)` and use
  `slog.InfoContext(ctx, ...)`. The HTTP layer's RequestID middleware
  already attaches the request ID to the context; logged events
  inherit it without additional plumbing

`slog.Default()` is used at the wiring site (`main.go`) only because
that file owns the global logger setup. Subscribers themselves never
call `slog.Default()`.

## Consequences

- New events (`domain/events/`) are pure value types with no
  dependencies on infrastructure. Adding one is ~25 lines.
- Adding a new subscriber for an existing event is purely additive:
  one new file in `infrastructure/`, one `Subscribe` line in
  `main.go`. No use case change, no test rewrite.
- Use case tests need a fake `EventPublisher` (a slice-backed recorder
  is enough) instead of a fake `Notifier` per concern. The fake is
  shared across all event-publishing use cases.
- A subscriber that panics is NOT recovered by the publisher — the
  panic propagates to the calling use case. Subscribers must not
  panic; if a recovery wrapper becomes necessary, it is added to the
  publisher in a follow-up.
- A subscriber that blocks blocks the use case. Subscribers that need
  to do I/O (HTTP calls, network) must use a tight timeout on the
  passed `ctx`.

## Alternatives considered and rejected

- **Direct `Notifier` port injected into the use case.** Works for one
  transport but does not scale: each new concern (metrics, audit,
  lockout) adds a constructor parameter and a new fake in tests. The
  Discord PR #55 originally proposed this and was reworked to use
  Domain Events instead.
- **`Notifier` factory** producing transport-specific instances per
  call. YAGNI for one transport (Discord), and once we have N
  transports the factory is just a degraded subscriber registry.
- **Async publisher with goroutines** for fire-and-forget delivery.
  Rejected for SMO's scale: the synchronous variant is simpler and
  the call sites are not in tight loops. An async variant can be added
  as a second adapter implementing the same port without changing the
  domain.
- **External message bus** (Kafka, NATS, Redis Streams). Massive
  overkill for a single-process app. The `EventPublisher` port leaves
  this open as a future implementation if the architecture grows.
