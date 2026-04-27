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

func newHealthRouter(db handlers.DBPinger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handlers.NewHealthHandler(db).Register(router)
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
