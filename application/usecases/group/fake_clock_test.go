package group

import "time"

// fakeClock is a deterministic implementation of the Clock port that
// always returns the same fixed time. Tests use it to make assertions
// about created_at timestamps without depending on real-time.
type fakeClock struct {
	now time.Time
}

func newFakeClock(now time.Time) *fakeClock {
	return &fakeClock{now: now}
}

func (c *fakeClock) Now() time.Time {
	return c.now
}
