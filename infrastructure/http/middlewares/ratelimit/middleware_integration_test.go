//go:build integration

package ratelimit_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	rdb "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	cacheredis "github.com/ianadou/smo/infrastructure/cache/redis"
	"github.com/ianadou/smo/infrastructure/http/middlewares/ratelimit"
)

// init disables the Ryuk reaper container before any test in this
// package starts.
//
// Ryuk fails to start on Fedora 43 + Docker 29 (known upstream bug),
// causing every testcontainers Run() call to error with "container
// is not running". Disabling Ryuk works around the local issue; the
// per-test t.Cleanup(_ = container.Terminate(ctx)) handles the
// cleanup that Ryuk would otherwise have done.
//
// Mirrors the same workaround used in the postgres repository, redis
// cache and cmd/server integration tests.
func init() {
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

// startRedis spins a fresh Redis testcontainer for the calling test.
// Each test gets its own container so test 7 (Redis-down runtime) can
// terminate the container mid-test without polluting other tests.
// Pinned to redis:7.4-alpine to avoid CI flakiness if the upstream
// :7-alpine tag drifts to an incompatible patch.
func startRedis(t *testing.T) (*rdb.Client, testcontainers.Container) {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7.4-alpine")
	if err != nil {
		t.Fatalf("start redis: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	url, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("redis url: %v", err)
	}
	client, err := cacheredis.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client, container
}

// router builds a tiny Gin engine with the rate-limit middleware
// attached and a single test route. Tests parameterize the route +
// config so they can simulate any policy.
func newRouter(t *testing.T, client *rdb.Client, cfg ratelimit.Config, route string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		t.Fatalf("trusted proxies: %v", err)
	}
	r.Use(ratelimit.New(client, cfg).Middleware())
	r.POST(route, func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

// retryAfterSeconds parses the Retry-After header on a 429 response.
// Used by tests that assert the value is in the expected range.
func retryAfterSeconds(t *testing.T, rec *httptest.ResponseRecorder) int {
	t.Helper()
	raw := rec.Header().Get("Retry-After")
	if raw == "" {
		t.Fatalf("expected Retry-After header on 429, got none")
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("Retry-After must be integer seconds, got %q (%v)", raw, err)
	}
	return n
}

func send(router *gin.Engine, route, xForwardedFor string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, route, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	if xForwardedFor != "" {
		req.Header.Set("X-Forwarded-For", xForwardedFor)
	}
	req.RemoteAddr = "127.0.0.1:1234" // simulate request from trusted proxy
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// 1. Sous la limite → 200 (204 here, the test handler returns NoContent).

func TestRateLimit_BelowLimit_AllRequestsPass(t *testing.T) {
	client, _ := startRedis(t)
	const route = "/probe"
	const limit = 5
	cfg := ratelimit.Config{route: {Limit: limit, Window: time.Minute}}
	router := newRouter(t, client, cfg, route)

	for i := 0; i < limit-1; i++ {
		rec := send(router, route, "1.2.3.4")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected 204, got %d", i+1, rec.Code)
		}
	}
}

// 2. À la limite exacte → 200 (Nème requête).

func TestRateLimit_AtLimit_LastRequestStillPasses(t *testing.T) {
	client, _ := startRedis(t)
	const route = "/probe"
	const limit = 3
	cfg := ratelimit.Config{route: {Limit: limit, Window: time.Minute}}
	router := newRouter(t, client, cfg, route)

	for i := 0; i < limit; i++ {
		rec := send(router, route, "1.2.3.4")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected 204, got %d (the Nth must still pass)", i+1, rec.Code)
		}
	}
}

// 3. Au-dessus → 429 + Retry-After valide.

func TestRateLimit_AboveLimit_Returns429WithRetryAfter(t *testing.T) {
	client, _ := startRedis(t)
	const route = "/probe"
	const limit = 2
	cfg := ratelimit.Config{route: {Limit: limit, Window: 30 * time.Second}}
	router := newRouter(t, client, cfg, route)

	for i := 0; i < limit; i++ {
		_ = send(router, route, "1.2.3.4")
	}
	rec := send(router, route, "1.2.3.4")

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on N+1, got %d", rec.Code)
	}
	retryAfter := retryAfterSeconds(t, rec)
	if retryAfter <= 0 || retryAfter > 30 {
		t.Errorf("Retry-After should be in (0, 30] seconds for a 30s window, got %d", retryAfter)
	}
}

// 4. Reset après expiration de la window → 200.

func TestRateLimit_WindowResets_AfterExpiry(t *testing.T) {
	client, _ := startRedis(t)
	const route = "/probe"
	const limit = 1
	const window = 2 * time.Second
	cfg := ratelimit.Config{route: {Limit: limit, Window: window}}
	router := newRouter(t, client, cfg, route)

	if rec := send(router, route, "1.2.3.4"); rec.Code != http.StatusNoContent {
		t.Fatalf("first request: expected 204, got %d", rec.Code)
	}
	if rec := send(router, route, "1.2.3.4"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request inside window: expected 429, got %d", rec.Code)
	}

	time.Sleep(window + 500*time.Millisecond)

	if rec := send(router, route, "1.2.3.4"); rec.Code != http.StatusNoContent {
		t.Fatalf("third request after window expiry: expected 204, got %d", rec.Code)
	}
}

// 5. IPs différentes → compteurs séparés.

func TestRateLimit_DifferentIPs_HaveIndependentCounters(t *testing.T) {
	client, _ := startRedis(t)
	const route = "/probe"
	const limit = 1
	cfg := ratelimit.Config{route: {Limit: limit, Window: time.Minute}}
	router := newRouter(t, client, cfg, route)

	if rec := send(router, route, "1.1.1.1"); rec.Code != http.StatusNoContent {
		t.Fatalf("ip1 first request: %d", rec.Code)
	}
	if rec := send(router, route, "1.1.1.1"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("ip1 second request: expected 429, got %d", rec.Code)
	}
	// IP 2 must be unaffected.
	if rec := send(router, route, "2.2.2.2"); rec.Code != http.StatusNoContent {
		t.Errorf("ip2 first request: expected 204 (independent counter), got %d", rec.Code)
	}
}

// 6. Routes différentes → compteurs séparés.

func TestRateLimit_DifferentRoutes_HaveIndependentCounters(t *testing.T) {
	client, _ := startRedis(t)
	cfg := ratelimit.Config{
		"/probe-a": {Limit: 1, Window: time.Minute},
		"/probe-b": {Limit: 1, Window: time.Minute},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})
	r.Use(ratelimit.New(client, cfg).Middleware())
	r.POST("/probe-a", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.POST("/probe-b", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	if rec := send(r, "/probe-a", "1.2.3.4"); rec.Code != http.StatusNoContent {
		t.Fatalf("probe-a first: %d", rec.Code)
	}
	if rec := send(r, "/probe-a", "1.2.3.4"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("probe-a second: expected 429, got %d", rec.Code)
	}
	// Same IP, different route: independent counter.
	if rec := send(r, "/probe-b", "1.2.3.4"); rec.Code != http.StatusNoContent {
		t.Errorf("probe-b first: expected 204 (independent counter), got %d", rec.Code)
	}
}

// 7. Redis down runtime → allow + log WARN.

func TestRateLimit_RedisDown_AtRuntime_AllowsRequest(t *testing.T) {
	client, container := startRedis(t)
	const route = "/probe"
	cfg := ratelimit.Config{route: {Limit: 1, Window: time.Minute}}
	router := newRouter(t, client, cfg, route)
	ctx := context.Background()

	// Verify the limit works while Redis is up.
	if rec := send(router, route, "1.2.3.4"); rec.Code != http.StatusNoContent {
		t.Fatalf("warmup: %d", rec.Code)
	}
	if rec := send(router, route, "1.2.3.4"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("warmup limit: expected 429, got %d", rec.Code)
	}

	// Kill Redis.
	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("terminate redis: %v", err)
	}

	// New request must pass through (soft fail).
	if rec := send(router, route, "9.9.9.9"); rec.Code != http.StatusNoContent {
		t.Errorf("post-kill request: expected 204 (soft fail), got %d", rec.Code)
	}
}

// 8. Trusted proxies → IP du X-Forwarded-For est utilisée.

func TestRateLimit_UsesXForwardedFor_WhenBehindTrustedProxy(t *testing.T) {
	client, _ := startRedis(t)
	const route = "/probe"
	const limit = 1
	cfg := ratelimit.Config{route: {Limit: limit, Window: time.Minute}}
	router := newRouter(t, client, cfg, route)

	// IP A maxes out via X-Forwarded-For.
	if rec := send(router, route, "10.20.30.40"); rec.Code != http.StatusNoContent {
		t.Fatalf("ip A first: %d", rec.Code)
	}
	if rec := send(router, route, "10.20.30.40"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("ip A second: expected 429, got %d", rec.Code)
	}
	// IP B (different X-Forwarded-For) is still under its own limit.
	// Same RemoteAddr (127.0.0.1) so this only works if Gin uses the
	// X-Forwarded-For header, which it does only because we configured
	// SetTrustedProxies(["127.0.0.1"]).
	if rec := send(router, route, "50.60.70.80"); rec.Code != http.StatusNoContent {
		t.Errorf("ip B first: expected 204 (X-Forwarded-For honored), got %d", rec.Code)
	}
}
