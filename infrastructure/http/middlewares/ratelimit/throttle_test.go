package ratelimit

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// captureSlog redirects slog.Default() to the returned buffer for the
// duration of the test. The original default logger is restored on
// cleanup.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })
	return &buf
}

func countWarnLines(buf *bytes.Buffer) int {
	return strings.Count(buf.String(), `"msg":"rate limit redis unavailable`)
}

// TestLimiter_MaybeWarnRedisDown_LogsOnceWithinThrottleWindow verifies
// the cold-path log throttling: many calls in quick succession should
// produce at most one WARN line per warnThrottle window.
func TestLimiter_MaybeWarnRedisDown_LogsOnceWithinThrottleWindow(t *testing.T) {
	buf := captureSlog(t)
	l := &Limiter{}
	err := errors.New("simulated redis down")
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		l.maybeWarnRedisDown(ctx, err)
	}

	if got := countWarnLines(buf); got != 1 {
		t.Errorf("expected exactly 1 WARN line within throttle window, got %d (output: %s)", got, buf.String())
	}
}

// TestLimiter_MaybeWarnRedisDown_LogsAgainAfterThrottleExpires verifies
// that the throttle does NOT silence the WARN forever. We advance the
// internal lastWarnAt manually rather than sleep warnThrottle (1 min)
// so the test stays fast.
func TestLimiter_MaybeWarnRedisDown_LogsAgainAfterThrottleExpires(t *testing.T) {
	buf := captureSlog(t)
	l := &Limiter{}
	err := errors.New("simulated redis down")
	ctx := context.Background()

	l.maybeWarnRedisDown(ctx, err)

	// Simulate that the previous WARN happened well outside the
	// throttle window.
	l.warnMu.Lock()
	l.lastWarnAt = time.Now().Add(-2 * warnThrottle)
	l.warnMu.Unlock()

	l.maybeWarnRedisDown(ctx, err)

	if got := countWarnLines(buf); got != 2 {
		t.Errorf("expected 2 WARN lines (one per window), got %d (output: %s)", got, buf.String())
	}
}

// TestLimiter_MaybeWarnRedisDown_FirstCallAlwaysLogs guards the zero-
// value case: a fresh Limiter has lastWarnAt set to time.Time{}, and
// the throttle MUST not silence the very first WARN (otherwise an
// outage at startup would log nothing for a full minute).
func TestLimiter_MaybeWarnRedisDown_FirstCallAlwaysLogs(t *testing.T) {
	buf := captureSlog(t)
	l := &Limiter{}

	l.maybeWarnRedisDown(context.Background(), errors.New("redis down"))

	if got := countWarnLines(buf); got != 1 {
		t.Errorf("expected first WARN to log, got %d", got)
	}
}
