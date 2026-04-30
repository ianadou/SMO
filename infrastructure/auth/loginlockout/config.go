package loginlockout

import "time"

// Config holds the policy parameters for the lockout tracker. The
// defaults below are deliberately conservative and align with the
// values agreed in issue #49 / ADR 0007.
type Config struct {
	// MaxFailures is the number of consecutive failed attempts within
	// FailureWindow that triggers a lockout. Counter resets on a
	// successful login or when FailureWindow elapses.
	MaxFailures int

	// FailureWindow is the rolling TTL on the failure counter. A
	// failure ages out and stops counting once this duration elapses
	// since the FIRST failure of the current window.
	FailureWindow time.Duration

	// LockoutDuration is how long the lockout flag stays set after
	// MaxFailures is crossed. The flag carries its own TTL and
	// automatically clears once it expires.
	LockoutDuration time.Duration
}

// DefaultConfig returns the production defaults: 5 failures in 15
// minutes triggers a 15-minute lockout. Aligned with the per-IP
// rate-limit policy in infrastructure/http/middlewares/ratelimit so
// the two layers reinforce each other instead of fighting.
func DefaultConfig() Config {
	return Config{
		MaxFailures:     5,
		FailureWindow:   15 * time.Minute,
		LockoutDuration: 15 * time.Minute,
	}
}
