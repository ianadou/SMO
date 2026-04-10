package ports

import "time"

// Clock returns the current time.
//
// Use cases inject this port instead of calling time.Now() directly so
// that tests can pass a fake clock returning a fixed timestamp. This
// makes assertions deterministic and avoids flaky tests caused by
// real-time drift between Now() calls.
type Clock interface {
	// Now returns the current time according to this clock.
	Now() time.Time
}
