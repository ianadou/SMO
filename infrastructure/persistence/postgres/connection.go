package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Default retry configuration for the initial database connection.
// These values are intentionally hardcoded: if the database is not
// reachable within ~30 seconds on startup, something is clearly wrong
// and failing fast is the right behavior.
const (
	connectMaxAttempts = 10
	connectRetryDelay  = 3 * time.Second
)

// Connect opens a pgxpool.Pool to the given connection string and
// retries on failure until either the pool responds to Ping() or the
// retry budget is exhausted.
//
// This retry loop matters in Docker Compose: even with healthchecks,
// there can be a brief window where Postgres accepts TCP connections
// but is not yet ready to answer queries. Retrying a few times with
// a short delay smooths over that window.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: build pool: %w", err)
	}

	var lastPingErr error
	for attempt := 1; attempt <= connectMaxAttempts; attempt++ {
		pingErr := pool.Ping(ctx)
		if pingErr == nil {
			return pool, nil
		}
		lastPingErr = pingErr

		if attempt < connectMaxAttempts {
			select {
			case <-ctx.Done():
				pool.Close()
				return nil, fmt.Errorf("postgres connect: context cancelled during retries: %w", ctx.Err())
			case <-time.After(connectRetryDelay):
			}
		}
	}

	pool.Close()
	return nil, fmt.Errorf(
		"postgres connect: ping failed after %d attempts: %w",
		connectMaxAttempts, lastPingErr,
	)
}
