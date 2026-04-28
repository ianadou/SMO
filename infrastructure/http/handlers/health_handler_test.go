package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/infrastructure/http/handlers"
)

type stubDBPinger struct {
	err error
}

func (s stubDBPinger) Ping(_ context.Context) error { return s.err }

type stubCachePinger struct {
	err error
}

func (s stubCachePinger) Ping(_ context.Context) error { return s.err }

// newHealthRouter builds a router with a HealthHandler whose cache is
// disabled (matches state 1 of ADR 0002). Tests that need a cache
// pinger build their own router with newHealthRouterWithCache.
func newHealthRouter(db handlers.DBPinger) *gin.Engine {
	return newHealthRouterWithCache(db, nil)
}

func newHealthRouterWithCache(db handlers.DBPinger, cache handlers.CachePinger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handlers.NewHealthHandler(db, cache).Register(router)
	return router
}

func TestHealthHandler_Live_Returns200_AlwaysAlive(t *testing.T) {
	t.Parallel()
	// A pinger that errors should not affect Live.
	router := newHealthRouter(stubDBPinger{err: errors.New("postgres is on fire")})

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "alive" {
		t.Errorf("expected status 'alive', got %v", body["status"])
	}
	if _, ok := body["timestamp"]; !ok {
		t.Errorf("expected timestamp field, got %v", body)
	}
}

func TestHealthHandler_Ready_Returns200_WhenDBPingSucceeds(t *testing.T) {
	t.Parallel()
	router := newHealthRouter(stubDBPinger{err: nil})

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "ready" {
		t.Errorf("expected status 'ready', got %v", body["status"])
	}
	if body["database"] != "ok" {
		t.Errorf("expected database 'ok', got %v", body["database"])
	}
	// No cache wired → must report "disabled", not "ok" (would lie)
	// nor "down" (would alarm).
	if body["cache"] != "disabled" {
		t.Errorf("expected cache 'disabled' when no pinger wired, got %v", body["cache"])
	}
}

func TestHealthHandler_Ready_Returns200_WhenCacheIsHealthy(t *testing.T) {
	t.Parallel()
	router := newHealthRouterWithCache(stubDBPinger{err: nil}, stubCachePinger{err: nil})

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["cache"] != "ok" {
		t.Errorf("expected cache 'ok', got %v", body["cache"])
	}
	if body["status"] != "ready" {
		t.Errorf("expected status 'ready', got %v", body["status"])
	}
}

func TestHealthHandler_Ready_Returns200_Degraded_WhenCacheIsDown(t *testing.T) {
	t.Parallel()
	// Cache is a soft dependency (ADR 0002): a Redis outage degrades
	// performance but the app keeps serving from Postgres. The probe
	// must keep returning 200 so the load balancer doesn't drain a
	// fully functional instance.
	router := newHealthRouterWithCache(
		stubDBPinger{err: nil},
		stubCachePinger{err: errors.New("redis timeout")},
	)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on cache-down (DB still ok), got %d (body=%s)",
			rec.Code, rec.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["cache"] != "down" {
		t.Errorf("expected cache 'down', got %v", body["cache"])
	}
	if body["status"] != "degraded" {
		t.Errorf("expected status 'degraded', got %v", body["status"])
	}
}

func TestHealthHandler_Ready_Returns503_WhenDBPingFails(t *testing.T) {
	t.Parallel()
	router := newHealthRouter(stubDBPinger{err: errors.New("connection refused")})

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "degraded" {
		t.Errorf("expected status 'degraded', got %v", body["status"])
	}
	if body["database"] != "down" {
		t.Errorf("expected database 'down', got %v", body["database"])
	}
}
