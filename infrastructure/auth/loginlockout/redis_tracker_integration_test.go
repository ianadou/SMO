//go:build integration

package loginlockout_test

import (
	"context"
	"os"
	"testing"
	"time"

	rdb "github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/ianadou/smo/infrastructure/auth/loginlockout"
	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
)

// init disables the Ryuk reaper container before any test in this
// package starts. Mirrors the same workaround used elsewhere in the
// repo for the Fedora 43 + Docker 29 testcontainers bug. Per-test
// t.Cleanup() takes over the cleanup Ryuk would have done.
func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

func startRedis(t *testing.T) *rdb.Client {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("start redis container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	url, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get redis URL: %v", err)
	}

	client, err := cacheredis.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect to redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return client
}

func newTrackerWithThreshold(client *rdb.Client, maxFailures int) *loginlockout.RedisTracker {
	return loginlockout.NewRedisTracker(client, loginlockout.Config{
		MaxFailures:     maxFailures,
		FailureWindow:   time.Minute,
		LockoutDuration: time.Minute,
	})
}

func TestRedisTracker_IsLocked_ReturnsFalse_WhenNoFailuresRecorded(t *testing.T) {
	client := startRedis(t)
	tracker := newTrackerWithThreshold(client, 5)
	ctx := context.Background()

	locked, err := tracker.IsLocked(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if locked {
		t.Errorf("fresh email must not be locked, got locked=true")
	}
}

func TestRedisTracker_LocksAccount_AfterReachingThreshold(t *testing.T) {
	client := startRedis(t)
	tracker := newTrackerWithThreshold(client, 3)
	ctx := context.Background()
	email := "alice@example.com"

	for i := 1; i <= 2; i++ {
		if err := tracker.RecordFailure(ctx, email); err != nil {
			t.Fatalf("RecordFailure %d: %v", i, err)
		}
		locked, _ := tracker.IsLocked(ctx, email)
		if locked {
			t.Fatalf("locked too early after %d failures (threshold 3)", i)
		}
	}

	if err := tracker.RecordFailure(ctx, email); err != nil {
		t.Fatalf("third RecordFailure: %v", err)
	}

	locked, err := tracker.IsLocked(ctx, email)
	if err != nil {
		t.Fatalf("IsLocked: %v", err)
	}
	if !locked {
		t.Errorf("expected lockout after 3 failures (threshold 3), got locked=false")
	}
}

func TestRedisTracker_RecordSuccess_ClearsCountersAndLockout(t *testing.T) {
	client := startRedis(t)
	tracker := newTrackerWithThreshold(client, 2)
	ctx := context.Background()
	email := "alice@example.com"

	_ = tracker.RecordFailure(ctx, email)
	_ = tracker.RecordFailure(ctx, email)
	locked, _ := tracker.IsLocked(ctx, email)
	if !locked {
		t.Fatalf("setup precondition: account should be locked after 2 failures")
	}

	if err := tracker.RecordSuccess(ctx, email); err != nil {
		t.Fatalf("RecordSuccess: %v", err)
	}

	locked, err := tracker.IsLocked(ctx, email)
	if err != nil {
		t.Fatalf("IsLocked after success: %v", err)
	}
	if locked {
		t.Errorf("RecordSuccess must clear lockout, got locked=true")
	}

	// Replay: a single failure after the reset must NOT immediately
	// re-lock the account, proving the failure counter was also cleared
	// (the threshold is 2, so 1 fresh failure must be insufficient).
	_ = tracker.RecordFailure(ctx, email)
	locked, _ = tracker.IsLocked(ctx, email)
	if locked {
		t.Errorf("counter must reset on success; one fresh failure should not re-lock with threshold 2")
	}
}

func TestRedisTracker_NormalizesEmailCase(t *testing.T) {
	client := startRedis(t)
	tracker := newTrackerWithThreshold(client, 2)
	ctx := context.Background()

	_ = tracker.RecordFailure(ctx, "alice@example.com")
	_ = tracker.RecordFailure(ctx, "Alice@Example.COM")

	locked, err := tracker.IsLocked(ctx, "ALICE@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("IsLocked: %v", err)
	}
	if !locked {
		t.Errorf("case variants must hit the same bucket; expected locked, got locked=false")
	}
}

func TestRedisTracker_FailsOpen_WhenClientIsClosed(t *testing.T) {
	client := startRedis(t)
	tracker := newTrackerWithThreshold(client, 5)
	ctx := context.Background()

	_ = client.Close()

	locked, err := tracker.IsLocked(ctx, "alice@example.com")
	if err != nil {
		t.Errorf("IsLocked must swallow Redis errors (fail-open), got %v", err)
	}
	if locked {
		t.Errorf("IsLocked must report not-locked on Redis error (fail-open), got locked=true")
	}

	if err := tracker.RecordFailure(ctx, "alice@example.com"); err != nil {
		t.Errorf("RecordFailure must swallow Redis errors (fail-open), got %v", err)
	}
	if err := tracker.RecordSuccess(ctx, "alice@example.com"); err != nil {
		t.Errorf("RecordSuccess must swallow Redis errors (fail-open), got %v", err)
	}
}
