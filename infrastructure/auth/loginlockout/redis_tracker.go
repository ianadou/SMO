package loginlockout

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"

	rdb "github.com/redis/go-redis/v9"
)

const (
	failedKeyPrefix = "smo:auth:failed:"
	lockedKeyPrefix = "smo:auth:locked:"

	// emailHashLen truncates the SHA-256 hex digest used in log fields.
	// 8 hex chars (32 bits) is enough to correlate a small number of
	// concurrent lockout events without exposing the email; collision
	// risk (~1 in 4 billion) is acceptable for telemetry, not for
	// authority decisions.
	emailHashLen = 8
)

// RedisTracker is the Redis-backed LoginAttemptTracker. Two keys per
// email:
//
//   - smo:auth:failed:<hashed_email>  INCR'd on each failure, EXPIRE
//     set to FailureWindow on first failure of a window.
//   - smo:auth:locked:<hashed_email>  SET to "1" with EX
//     LockoutDuration once MaxFailures is crossed.
//
// The hashed email is what lands in keys: this matches the slog
// behavior (we never want a plain email to leak into Redis MONITOR
// output, RDB snapshots, or replication logs). Normalization is
// strings.ToLower applied before hashing so the casing rule cannot be
// bypassed by alternating "Alice@x.com" / "alice@x.com".
//
// FAIL-OPEN POLICY (intentional, see ADR 0007):
// every Redis call is allowed to fail. IsLocked returns (false, nil)
// on error — a Redis hiccup never converts into a 401. RecordFailure
// and RecordSuccess return nil on error — the use case keeps going.
// All three log a structured WARN so operators can correlate
// degradation. Without this, Redis flapping would lock every user out
// of the product.
type RedisTracker struct {
	client *rdb.Client
	cfg    Config
}

// NewRedisTracker wraps a redis client with the given policy. The
// client must not be nil — callers select NoopTracker when Redis is
// disabled.
func NewRedisTracker(client *rdb.Client, cfg Config) *RedisTracker {
	return &RedisTracker{client: client, cfg: cfg}
}

// IsLocked reports whether the lockout flag is currently set for
// email. Errors are logged and downgraded to (false, nil) — see the
// fail-open policy on the type.
func (t *RedisTracker) IsLocked(ctx context.Context, email string) (bool, error) {
	hashed := hashEmail(email)
	key := lockedKeyPrefix + hashed

	n, err := t.client.Exists(ctx, key).Result()
	if err != nil {
		t.warnDegraded(ctx, "IsLocked", hashed, err)
		return false, nil
	}
	return n > 0, nil
}

// RecordFailure increments the failure counter and, when MaxFailures
// is crossed, sets the lockout flag. Errors are logged and swallowed.
func (t *RedisTracker) RecordFailure(ctx context.Context, email string) error {
	hashed := hashEmail(email)
	failedKey := failedKeyPrefix + hashed

	count, err := t.client.Incr(ctx, failedKey).Result()
	if err != nil {
		t.warnDegraded(ctx, "RecordFailure", hashed, err)
		return nil
	}
	if count == 1 {
		if expErr := t.client.Expire(ctx, failedKey, t.cfg.FailureWindow).Err(); expErr != nil {
			t.warnDegraded(ctx, "RecordFailure.Expire", hashed, expErr)
			return nil
		}
	}
	if count >= int64(t.cfg.MaxFailures) {
		lockedKey := lockedKeyPrefix + hashed
		if setErr := t.client.Set(ctx, lockedKey, "1", t.cfg.LockoutDuration).Err(); setErr != nil {
			t.warnDegraded(ctx, "RecordFailure.Set", hashed, setErr)
			return nil
		}
		slog.WarnContext(ctx, "account locked due to repeated failed login attempts",
			slog.String("email_hash", hashed),
			slog.Int64("failure_count", count),
			slog.Int("lockout_duration_seconds", int(t.cfg.LockoutDuration.Seconds())),
			slog.String("tracker_backend", "redis"),
		)
	}
	return nil
}

// RecordSuccess clears both counters for email. Errors are logged and
// swallowed.
func (t *RedisTracker) RecordSuccess(ctx context.Context, email string) error {
	hashed := hashEmail(email)
	if err := t.client.Del(
		ctx,
		failedKeyPrefix+hashed,
		lockedKeyPrefix+hashed,
	).Err(); err != nil {
		t.warnDegraded(ctx, "RecordSuccess", hashed, err)
	}
	return nil
}

// warnDegraded emits a structured warning with the same shape across
// all three operations, so log queries can group by operation. It
// also redacts the Redis error to its message — the full client
// config (addresses, etc.) must never land in a log.
func (t *RedisTracker) warnDegraded(ctx context.Context, op, hashed string, err error) {
	slog.WarnContext(ctx, "login attempt tracker degraded, failing open",
		slog.String("operation", op),
		slog.String("email_hash", hashed),
		slog.String("error", redactRedisError(err)),
		slog.String("tracker_backend", "redis"),
	)
}

func redactRedisError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "context canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "context deadline exceeded"
	}
	return err.Error()
}

// hashEmail returns the first emailHashLen hex chars of SHA-256 over
// the lowercased email. SHA-256 (not bcrypt) because this is a
// telemetry correlation token, not a stored credential — the cost of
// bcrypt would dominate a hot path that runs on every login attempt.
func hashEmail(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(email)))
	return hex.EncodeToString(sum[:])[:emailHashLen]
}
