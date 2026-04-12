package invitation

import "time"

type fakeClock struct{ now time.Time }

func newFakeClock(now time.Time) *fakeClock { return &fakeClock{now: now} }

func (c *fakeClock) Now() time.Time { return c.now }
