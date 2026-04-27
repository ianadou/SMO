package redis_test

import (
	"context"
	"testing"
	"time"

	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
)

func TestConnect_ReturnsNilClient_WhenURLIsEmpty(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client, err := cacheredis.Connect(ctx, "")
	if err != nil {
		t.Errorf("expected no error for empty URL, got %v", err)
	}
	if client != nil {
		t.Errorf("expected nil client for empty URL, got %v", client)
	}
}

func TestConnect_ReturnsError_WhenURLIsMalformed(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client, err := cacheredis.Connect(ctx, "not-a-valid-url")

	if err == nil {
		t.Errorf("expected error for malformed URL, got nil")
	}
	if client != nil {
		t.Errorf("expected nil client when error is returned, got %v", client)
	}
}

func TestConnect_ReturnsError_WhenRedisIsUnreachable(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	// Port 1 is reserved and refuses connections; the dial fails fast.
	client, err := cacheredis.Connect(ctx, "redis://127.0.0.1:1/0")

	if err == nil {
		t.Errorf("expected error for unreachable Redis, got nil")
		_ = client.Close()
	}
	if client != nil {
		t.Errorf("expected nil client when Redis is unreachable, got %v", client)
	}
}
