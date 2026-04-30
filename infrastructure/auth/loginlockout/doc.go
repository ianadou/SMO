// Package loginlockout implements LoginAttemptTracker (the domain port
// from domain/ports/login_attempt_tracker.go).
//
// Two adapters live here:
//
//   - RedisTracker stores per-account failure counters and lockout
//     flags in Redis with TTLs, so cleanup is automatic and no schema
//     migration is needed.
//   - NoopTracker is the fallback used when Redis is disabled
//     (REDIS_URL empty per ADR 0002). It always reports "not locked",
//     never errors, and records nothing — the login flow runs without
//     account-level lockout in that mode.
//
// Both adapters are fail-open by design: a Redis error degrades to
// "let the login through" and emits a WARN, never a 5xx. See ADR 0007
// for the rationale (avoid turning a Redis hiccup into a global auth
// outage).
package loginlockout
