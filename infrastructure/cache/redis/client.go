package redis

import (
	"context"
	"fmt"
	"time"

	rdb "github.com/redis/go-redis/v9"
)

const connectTimeout = 5 * time.Second

// Connect interprets the REDIS_URL value and returns either:
//
//   - (nil, nil)        when url is empty: cache is disabled, callers
//     should NOT wrap repositories. Logs at INFO level upstream.
//   - (*Client, nil)    when url is set and Redis is reachable.
//   - (nil, error)      when url is set but Redis is unreachable, the
//     URL is malformed, or any other config-time failure. Caller
//     should treat this as fatal.
//
// This 3-state contract is the foundation of ADR 0002.
func Connect(ctx context.Context, url string) (*rdb.Client, error) {
	if url == "" {
		return nil, nil
	}

	opts, err := rdb.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
	}

	client := rdb.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if pingErr := client.Ping(pingCtx).Err(); pingErr != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect to redis: %w", pingErr)
	}

	return client, nil
}

// Pinger adapts a *redis.Client into a small Ping(ctx) error contract,
// so callers (e.g. the health handler) don't need to know about
// *redis.StatusCmd or the go-redis package shape.
//
// Callers should pass nil instead of a Pinger when the cache is
// disabled by configuration (ADR 0002 state 1, REDIS_URL unset). The
// health handler treats a nil pinger as "cache disabled", which is a
// distinct state from "cache configured but unreachable".
type Pinger struct {
	client *rdb.Client
}

// NewPinger wraps a non-nil *redis.Client into a context-only pinger.
// Passing a nil client is a programming error; the caller should pass
// nil to the health handler instead of constructing a Pinger.
func NewPinger(client *rdb.Client) *Pinger {
	return &Pinger{client: client}
}

// Ping reports whether Redis answers a PING within the deadline of ctx.
func (p *Pinger) Ping(ctx context.Context) error {
	if err := p.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}
