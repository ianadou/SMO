//go:build integration

package redis_test

import (
	"context"
	"strings"
	"testing"

	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
)

// TestPinger_Ping_ReturnsNil_WhenRedisIsReachable is the happy-path
// guard: the small Pinger adapter wired into the health handler must
// pass through to a successful PING.
func TestPinger_Ping_ReturnsNil_WhenRedisIsReachable(t *testing.T) {
	client, _ := startRedis(t)

	pinger := cacheredis.NewPinger(client)

	if err := pinger.Ping(context.Background()); err != nil {
		t.Errorf("expected nil for reachable Redis, got %v", err)
	}
}

// TestPinger_Ping_ReturnsError_WhenRedisIsDown asserts the wrapped
// error path. Terminating the container mid-test simulates a runtime
// outage; Ping must surface the failure so the readiness probe can
// flip cache="down". The wrapped error is what /health/ready uses to
// decide between "ok" and "down".
func TestPinger_Ping_ReturnsError_WhenRedisIsDown(t *testing.T) {
	client, container := startRedis(t)

	if err := container.Terminate(context.Background()); err != nil {
		t.Fatalf("terminate redis: %v", err)
	}

	pinger := cacheredis.NewPinger(client)

	err := pinger.Ping(context.Background())
	if err == nil {
		t.Fatalf("expected non-nil error after Redis termination, got nil")
	}
	// The wrapper prefixes the message with "redis ping:" so callers
	// (and logs) can identify the source.
	if !strings.Contains(err.Error(), "redis ping") {
		t.Errorf("expected error to mention 'redis ping', got %q", err.Error())
	}
}
