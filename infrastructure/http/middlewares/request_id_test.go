package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ianadou/smo/infrastructure/http/middlewares"
)

func newRouterWithRequestID() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middlewares.RequestID())
	router.GET("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"request_id": middlewares.RequestIDFromContext(c.Request.Context()),
		})
	})
	return router
}

func TestRequestID_GeneratesUUIDv4_WhenHeaderIsMissing(t *testing.T) {
	t.Parallel()
	router := newRouterWithRequestID()

	req := httptest.NewRequest(http.MethodGet, "/echo", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	got := rec.Header().Get(middlewares.RequestIDHeader)
	parsed, err := uuid.Parse(got)
	if err != nil {
		t.Fatalf("expected response header to be a valid UUID, got %q: %v", got, err)
	}
	if parsed.Version() != 4 {
		t.Errorf("expected UUID v4, got version %d", parsed.Version())
	}
	if rec.Header().Get(middlewares.RequestIDGeneratedHeader) != "true" {
		t.Errorf("expected X-Request-ID-Generated: true, got %q",
			rec.Header().Get(middlewares.RequestIDGeneratedHeader))
	}
}

func TestRequestID_PropagatesIncomingValidUUIDv4(t *testing.T) {
	t.Parallel()
	router := newRouterWithRequestID()

	incoming := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/echo", nil)
	req.Header.Set(middlewares.RequestIDHeader, incoming)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get(middlewares.RequestIDHeader); got != incoming {
		t.Errorf("expected propagated %q, got %q", incoming, got)
	}
	if rec.Header().Get(middlewares.RequestIDGeneratedHeader) == "true" {
		t.Errorf("expected no X-Request-ID-Generated header for valid incoming ID")
	}
}

func TestRequestID_RegeneratesWhenIncomingIsInvalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, header string
	}{
		{"non-uuid garbage", "not-a-uuid"},
		{"uuid v1", "00000000-0000-1000-8000-000000000000"},
		{"empty quoted", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			router := newRouterWithRequestID()

			req := httptest.NewRequest(http.MethodGet, "/echo", nil)
			if tc.header != "" {
				req.Header.Set(middlewares.RequestIDHeader, tc.header)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			got := rec.Header().Get(middlewares.RequestIDHeader)
			parsed, err := uuid.Parse(got)
			if err != nil || parsed.Version() != 4 {
				t.Errorf("expected fresh UUID v4, got %q (err=%v)", got, err)
			}
			if rec.Header().Get(middlewares.RequestIDGeneratedHeader) != "true" {
				t.Errorf("expected X-Request-ID-Generated: true for invalid incoming")
			}
		})
	}
}

func TestRequestIDFromContext_ReturnsEmpty_WhenMiddlewareNotApplied(t *testing.T) {
	t.Parallel()

	got := middlewares.RequestIDFromContext(httptest.NewRequest(http.MethodGet, "/", nil).Context())

	if got != "" {
		t.Errorf("expected empty string when middleware not applied, got %q", got)
	}
}
