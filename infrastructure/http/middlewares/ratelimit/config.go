package ratelimit

import "time"

// RouteSpec defines the rate-limit policy for a single route: at most
// Limit requests per IP within a Window-long fixed window.
type RouteSpec struct {
	Limit  int
	Window time.Duration
}

// Config maps Gin route patterns (the value of c.FullPath()) to their
// rate-limit policy. Routes not present in the map are not rate-limited.
type Config map[string]RouteSpec

// DefaultConfig returns the SMO production rate-limit policy. Values
// are hardcoded here for now; a follow-up PR can move them to env
// vars if operators need to tune without rebuilding.
//
// Login: deliberately conservative (5/15min, industry standard) until
// the account-level lockout from issue #49 lands. The IP throttle is
// only a complementary layer against brute-force.
func DefaultConfig() Config {
	return Config{
		"/api/v1/auth/login":         {Limit: 5, Window: 15 * time.Minute},
		"/api/v1/auth/register":      {Limit: 3, Window: time.Hour},
		"/api/v1/invitations/accept": {Limit: 10, Window: time.Minute},
	}
}
