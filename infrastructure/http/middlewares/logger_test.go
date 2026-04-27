package middlewares_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/infrastructure/http/middlewares"
)

func newRouterWithLogger(buf *bytes.Buffer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	router.Use(middlewares.RequestID())
	router.Use(middlewares.SLogLogger(logger))
	router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/boom", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}

func decodeLogLine(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("expected JSON log line, got %q: %v", string(raw), err)
	}
	return m
}

func TestSLogLogger_LogsCompletedRequest_WithExpectedFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	router := newRouterWithLogger(&buf)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if buf.Len() == 0 {
		t.Fatalf("expected a log line, got empty buffer")
	}
	line := decodeLogLine(t, buf.Bytes())

	if line["msg"] != "request completed" {
		t.Errorf("expected msg 'request completed', got %v", line["msg"])
	}
	if line["method"] != "GET" {
		t.Errorf("expected method GET, got %v", line["method"])
	}
	if line["path"] != "/ok" {
		t.Errorf("expected path /ok, got %v", line["path"])
	}
	if line["status"] != float64(200) {
		t.Errorf("expected status 200, got %v", line["status"])
	}
	if _, ok := line["duration_ms"]; !ok {
		t.Errorf("expected duration_ms field, got %v", line)
	}
	if id, _ := line["request_id"].(string); id == "" {
		t.Errorf("expected non-empty request_id, got %v", line["request_id"])
	}
}

func TestSLogLogger_UsesErrorLevel_For5xx(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	router := newRouterWithLogger(&buf)

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	line := decodeLogLine(t, buf.Bytes())
	if line["level"] != "ERROR" {
		t.Errorf("expected level ERROR for 500, got %v", line["level"])
	}
}

func TestSLogLogger_SkipsHealthEndpoint(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	router := newRouterWithLogger(&buf)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log line for /health, got %q", buf.String())
	}
}
