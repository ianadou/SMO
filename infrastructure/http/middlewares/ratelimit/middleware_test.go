package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/infrastructure/http/middlewares/ratelimit"
)

// Unit-level tests on the no-op path: a nil Redis client must produce
// a pass-through middleware without ever touching the network. This
// is the cache-disabled state from ADR 0002, scenario 9 in the test
// matrix promised at scope time.

func TestMiddleware_NilClient_PassesAllRequestsThrough(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ratelimit.New(nil, ratelimit.DefaultConfig()).Middleware())
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Hit the route many more times than the configured login limit.
	// All responses must be 204; none must be 429.
	const calls = 50
	for i := 0; i < calls; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("call %d: expected 204 (pass-through), got %d", i, rec.Code)
		}
	}
}

func TestDefaultConfig_DoesNotIncludeVotesOrReads(t *testing.T) {
	t.Parallel()

	cfg := ratelimit.DefaultConfig()

	for _, path := range []string{
		"/api/v1/votes",
		"/api/v1/groups/:id",
		"/api/v1/matches/:id",
	} {
		if _, found := cfg[path]; found {
			t.Errorf("%s should NOT be rate-limited; only login/register/accept-invitation are", path)
		}
	}
	for _, path := range []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/invitations/accept",
	} {
		if _, found := cfg[path]; !found {
			t.Errorf("%s must be rate-limited", path)
		}
	}
}
