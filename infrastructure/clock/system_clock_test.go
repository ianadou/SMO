package clock

import (
	"testing"
	"time"
)

func TestSystemClock_Now_ReturnsCurrentTime(t *testing.T) {
	t.Parallel()

	c := New()
	before := time.Now()
	got := c.Now()
	after := time.Now()

	// The returned time must lie within the wall-clock window we
	// observed around the call. A drift of more than a few seconds
	// would indicate a fundamentally broken clock implementation.
	if got.Before(before) || got.After(after) {
		t.Errorf("Now() = %v, want between %v and %v", got, before, after)
	}
}
