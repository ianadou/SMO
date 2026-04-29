package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/infrastructure/http/middlewares"
)

func newCORSRouter(t *testing.T, allowed []string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.CORS(allowed))
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"pong": true}) })
	r.POST("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"pong": true}) })
	return r
}

func TestCORS_AllowedOrigin_GetsAllowOriginHeader(t *testing.T) {
	t.Parallel()
	r := newCORSRouter(t, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("expected origin echo header, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected credentials true, got %q", got)
	}
}

func TestCORS_DisallowedOrigin_DoesNotSetAllowOriginHeader(t *testing.T) {
	t.Parallel()
	r := newCORSRouter(t, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("request must still succeed, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no allow-origin for disallowed origin, got %q", got)
	}
}

func TestCORS_NoOriginHeader_PassesThroughWithoutCORSHeaders(t *testing.T) {
	t.Parallel()
	r := newCORSRouter(t, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("non-browser request must not get CORS headers, got %q", got)
	}
}

func TestCORS_PreflightRequest_Returns204(t *testing.T) {
	t.Parallel()
	r := newCORSRouter(t, []string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 on preflight, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Errorf("expected allow-methods header on preflight, got empty")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Errorf("expected allow-headers header on preflight, got empty")
	}
}

func TestCORS_TrimsAndIgnoresEmptyEntries(t *testing.T) {
	t.Parallel()
	r := newCORSRouter(t, []string{"  http://localhost:3000  ", "", "  "})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("expected trimmed origin to match, got %q", got)
	}
}
