package clock

import "time"

// SystemClock returns the current system time. It implements the
// domain ports.Clock interface.
//
// In production code, this is the clock to inject. In tests, inject
// a fake clock that returns a predetermined time so assertions are
// deterministic.
type SystemClock struct{}

// New returns a new SystemClock. The constructor exists for symmetry
// with other adapters and to make the dependency explicit at the
// composition root.
func New() *SystemClock {
	return &SystemClock{}
}

// Now returns the current system time using time.Now().
func (c *SystemClock) Now() time.Time {
	return time.Now()
}
