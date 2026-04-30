# ADR 0007 — Account-level lockout on failed login

**Status:** Accepted (2026-04-30)

## Context

SMO already has two layers of defense on `POST /api/v1/auth/login`:

1. **Password hashing** (bcrypt, cost 12 — see ADR 0001 era code) makes
   any individual credential check expensive.
2. **Per-IP rate limit** (5 requests / 15 minutes — see
   `infrastructure/http/middlewares/ratelimit/config.go`) caps how
   fast a single IP can iterate.

Neither defends against an attacker rotating IPs across a botnet to
brute-force one specific account. A 1000-IP botnet under the per-IP
limit can attempt 5000 passwords per 15-minute window without any
single IP being throttled — that is enough to crack a weak password
on a high-value account in hours, not days.

The classic mitigation is a **per-account lockout**: count failures
keyed on the email being attacked, not on the source IP. Once a
threshold is crossed, freeze logins to that account for a cooldown
period regardless of where the attempts come from.

This ADR records the decision to add that lockout, the trade-offs
considered, and the operational invariants future maintainers should
preserve.

## Decision

### Per-account lockout in Redis with TTLs

A new domain port `LoginAttemptTracker` exposes three operations:

- `IsLocked(email)` — checked before the password compare.
- `RecordFailure(email)` — called after every failed attempt.
- `RecordSuccess(email)` — called after every successful login.

The production adapter is `loginlockout.RedisTracker`. It maintains
two keys per email:

- `smo:auth:failed:<hashed_email>` — INCR'd on each failure, EXPIRE
  set to `FailureWindow` on the first failure of the window.
- `smo:auth:locked:<hashed_email>` — SET with EX `LockoutDuration`
  once `MaxFailures` is crossed.

Defaults (`Config{}`): **5 failures within 15 minutes triggers a
15-minute lockout**. Aligned with the existing per-IP rate-limit
window so the two layers reinforce each other.

### Hashed email in keys and logs

Both the Redis key and the structured log field `email_hash` use
SHA-256 truncated to 8 hex chars (32 bits). This keeps the key
namespace and the log stream useful for correlation without putting
plain emails into Redis MONITOR output, RDB snapshots, replication
streams, or log aggregation pipelines. Collisions (~1 in 4 billion)
are acceptable for telemetry; this is not an authority decision.

SHA-256 is chosen over bcrypt deliberately: bcrypt would dominate the
hot path, and the goal here is correlation, not credential storage.

### Fail-open on infrastructure errors

If any Redis call returns an error, the tracker:

1. Logs a structured `WARN` ("login attempt tracker degraded, failing
   open") with the operation name and redacted error.
2. Returns the safe default: `IsLocked` returns `false`,
   `RecordFailure` and `RecordSuccess` return `nil`.

The use case treats a tracker error as a no-op and proceeds. The
explicit comment on `RedisTracker` and the matching policy on
`LoginAttemptTracker` (in `domain/ports/`) record that this is
deliberate, not an oversight.

### Indistinguishable response on lockout

`ErrAccountLocked` is a distinct domain error (so logs and metrics
can tell it apart from `ErrInvalidCredentials`), but the HTTP mapper
folds both into the same response: `401 Unauthorized` with body
`{"error": "invalid credentials"}`. A distinct "account locked"
response would tell an attacker that the email is valid AND under
attack — defeating the existing email-enumeration defense (see
`LoginOrganizerUseCase` comment).

### NoopTracker fallback

When `REDIS_URL` is empty, `cmd/server/main.go` wires
`loginlockout.NewNoopTracker()` instead of the Redis adapter. The
boot log emits a `WARN` so operators see they are running without
the lockout defense. This mirrors how the per-IP rate-limit middleware
degrades to pass-through in the same configuration (see ADR 0002).

## Alternatives rejected

### Per-IP global counter (e.g., 50 failures across all emails per IP)

Closes the "100 emails × 4 attempts each" bypass that the issue
describes. Rejected for **this PR** because:

- The trade-off with corporate NAT is not yet calibrated — a single
  office IP can legitimately serve 50+ employees.
- The per-IP rate limit on `/auth/login` already provides partial
  coverage for the worst slow cases.
- We have no real attack data to inform the threshold.

Left as future work; can be added as a second `RecordFailure` key
without restructuring the port.

### DB-backed lockout (Postgres column on `organizers`)

Rejected. The state is short-lived (15 minutes), high-write (every
failed login), and self-expiring — exactly the workload Redis with
TTL is built for. A Postgres column would force a migration, vacuum
churn from constant writes, and manual cleanup of stale lockouts.

### Fail-closed on Redis errors

Rejected. A Redis outage would convert into a global authentication
outage: every login would fail because the lockout check itself
errored. SMO is small enough that the availability cost of a
fail-closed posture outweighs the security benefit of "definitely no
brute force during a Redis outage". The ratelimit middleware made
the same call (see ADR 0002).

### Bcrypt-hashed email keys

Rejected. Bcrypt is designed to be slow; hashing the email on every
login attempt would put a deliberate slowdown on the hot path of an
authentication endpoint. SHA-256 truncation gives sufficient
unlinkability for a telemetry/key-namespace hash.

## Consequences

### Targeted DoS of a known account is possible

An attacker who knows a victim's email can lock them out for 15
minutes by submitting 5 wrong passwords. This is the classic trade-off
of any account-lockout mechanism. Acceptable for SMO at its scale; if
this becomes a real problem, mitigations include:

- Allowing the legitimate user to bypass via a "send me a reset link"
  flow that does not increment the counter.
- Switching to a CAPTCHA after N failures instead of a hard lock.
- Differentiating per-IP behavior (whitelist the user's last-known
  good IP).

None of these are needed for the MVP.

### Defense disappears silently if Redis goes down

This is the cost of the fail-open posture. The boot log and the
`WARN` on every degraded call provide visibility, and the per-IP
rate limit (also Redis-backed) flips to pass-through at the same
moment, so the system advertises its degraded state via metrics.
Operators must treat a sustained Redis outage as a security incident,
not just a performance one.

### Counter normalization is the load-bearing invariant

Every entry point to the tracker normalizes the email with
`strings.ToLower` before hashing. If a future caller forgets this,
case variants like `Alice@x.com` and `alice@x.com` would land in
different keys and the threshold would never trigger. The
integration test `TestRedisTracker_NormalizesEmailCase` is the
guard.

## Future work

- Per-IP global counter as a follow-up if real attack data shows the
  per-account-only posture is being bypassed at scale.
- Prometheus counter `auth_account_lockout_total` with a
  `tracker_backend` label, to make the noop-vs-redis split visible
  in dashboards.
- Optional admin tool to manually clear a lockout (reset link, ops
  unlock command) once the reset-password flow exists.
