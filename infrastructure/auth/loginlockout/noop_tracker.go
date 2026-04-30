package loginlockout

import "context"

// NoopTracker is the fallback used when Redis is disabled (empty
// REDIS_URL per ADR 0002). It satisfies LoginAttemptTracker but
// performs no bookkeeping: IsLocked always returns false, RecordFailure
// and RecordSuccess are no-ops.
//
// In this mode the system loses the per-account lockout defense and
// must rely solely on the per-IP rate limit (which itself is also a
// no-op without Redis). main.go logs the active backend at boot so
// operators see this degradation explicitly.
type NoopTracker struct{}

// NewNoopTracker constructs the no-op tracker.
func NewNoopTracker() *NoopTracker {
	return &NoopTracker{}
}

// IsLocked always returns false: without backing storage, no account
// can ever be flagged as locked.
func (NoopTracker) IsLocked(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// RecordFailure is a no-op.
func (NoopTracker) RecordFailure(_ context.Context, _ string) error {
	return nil
}

// RecordSuccess is a no-op.
func (NoopTracker) RecordSuccess(_ context.Context, _ string) error {
	return nil
}
