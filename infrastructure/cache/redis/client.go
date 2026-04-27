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
