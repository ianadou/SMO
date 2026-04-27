package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	rdb "github.com/redis/go-redis/v9"
)

// keyPrefix isolates rate-limit keys from other Redis users (cache
// entries from PR #37 use "smo:group:..." etc.).
const keyPrefix = "smo:ratelimit:"

// warnThrottle caps Redis-down WARN logs to one per minute to avoid
// flooding the log stream when Redis is unavailable for several
// seconds across hundreds of requests.
const warnThrottle = time.Minute

// Limiter is the stateful runtime returned by New. The struct exposes
// no methods directly; clients use it through the Middleware() Gin
// handler. Holding state on the limiter (rather than capturing it in
// a closure) makes the throttled WARN logic and any future metrics
// straightforward to test.
//
// warnMu protects lastWarnAt. A mutex was chosen over an atomic CAS
// because Redis-down is a cold path (rare): the few-nanosecond cost
// of a mutex acquisition is invisible compared to the readability
// gain. lastWarnAt is a time.Time so time.Since() can use the
// monotonic clock — immune to NTP wall-clock jumps that would break
// a unix-nanos comparison.
type Limiter struct {
	client     *rdb.Client
	config     Config
	warnMu     sync.Mutex
	lastWarnAt time.Time // zero value means we have never warned yet
}

// New constructs a Limiter. When client is nil (cache disabled state
// from ADR 0002), the limiter is in pass-through mode and Middleware()
// returns a no-op handler.
func New(client *rdb.Client, config Config) *Limiter {
	return &Limiter{client: client, config: config}
}

// Middleware returns the Gin handler that enforces the limiter's
// policy on each request. The handler is safe to attach globally on
// the router because it short-circuits to c.Next() on routes that
// have no spec in the config.
func (l *Limiter) Middleware() gin.HandlerFunc {
	if l.client == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		spec, found := l.config[c.FullPath()]
		if !found {
			c.Next()
			return
		}

		key := keyPrefix + c.FullPath() + ":" + c.ClientIP()
		count, ttl, err := l.incr(c.Request.Context(), key, spec.Window)
		if err != nil {
			l.maybeWarnRedisDown(c.Request.Context(), err)
			c.Next()
			return
		}

		if count > int64(spec.Limit) {
			retryAfter := ttl
			if retryAfter <= 0 {
				retryAfter = spec.Window
			}
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

// incr atomically increments the counter for key. If the counter was
// just created (count == 1), an EXPIRE is set so the window starts
// counting from the first request. Returns the post-increment count
// and the remaining TTL on the key, plus any Redis error.
func (l *Limiter) incr(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("ratelimit incr: %w", err)
	}
	if count == 1 {
		if expErr := l.client.Expire(ctx, key, window).Err(); expErr != nil {
			return 0, 0, fmt.Errorf("ratelimit expire: %w", expErr)
		}
	}
	ttl, err := l.client.TTL(ctx, key).Result()
	if err != nil {
		return count, 0, fmt.Errorf("ratelimit ttl: %w", err)
	}
	return count, ttl, nil
}

// maybeWarnRedisDown logs at most once per warnThrottle interval so a
// Redis outage that affects many requests does not flood the logs.
// Uses time.Since(lastWarnAt) which reads the monotonic clock, so the
// throttle is unaffected by wall-clock NTP corrections.
func (l *Limiter) maybeWarnRedisDown(ctx context.Context, err error) {
	l.warnMu.Lock()
	defer l.warnMu.Unlock()
	if !l.lastWarnAt.IsZero() && time.Since(l.lastWarnAt) < warnThrottle {
		return
	}
	l.lastWarnAt = time.Now()
	slog.WarnContext(ctx, "rate limit redis unavailable, allowing requests",
		slog.String("error", redactRedisError(err)))
}

// redactRedisError reduces a redis error to its message so we do not
// inadvertently log the full client config (addresses, etc.).
func redactRedisError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "context canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "context deadline exceeded"
	}
	return err.Error()
}
