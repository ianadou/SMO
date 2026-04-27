package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the HTTP header carrying the request identifier
	// in both directions (incoming for trace propagation from upstream
	// proxies, outgoing for client-side debugging).
	RequestIDHeader = "X-Request-ID"

	// RequestIDGeneratedHeader is set on the response when the server
	// generated a fresh request ID because the incoming X-Request-ID
	// was missing or invalid. Helps an integrator notice that their
	// trace propagation isn't working.
	RequestIDGeneratedHeader = "X-Request-ID-Generated"

	// requestIDGinKey is the gin.Context key under which the request ID
	// is stored. Handlers should prefer RequestIDFromContext over reading
	// the gin context directly: the context.Context value works in both
	// HTTP handlers and downstream use cases.
	requestIDGinKey = "request_id"
)

// requestIDContextKey is an unexported type used as the key for storing
// the request ID in context.Context. Using a typed key (rather than a
// raw string) avoids collisions with values from other packages.
type requestIDContextKey struct{}

// RequestID returns a Gin middleware that ensures every request carries
// a stable identifier. If the incoming request has a valid UUID v4 in
// the X-Request-ID header, it is propagated as-is. Otherwise a fresh
// UUID v4 is generated and the response carries an
// X-Request-ID-Generated: true header to signal this to the caller.
//
// The request ID is exposed both in the gin.Context (via "request_id")
// and in the standard context.Context, so downstream use cases can
// retrieve it via RequestIDFromContext without taking a Gin dependency.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		incoming := c.GetHeader(RequestIDHeader)
		generated := false

		id := incoming
		if !isValidUUIDv4(id) {
			id = uuid.NewString()
			generated = true
		}

		c.Set(requestIDGinKey, id)
		c.Header(RequestIDHeader, id)
		if generated {
			c.Header(RequestIDGeneratedHeader, "true")
		}

		ctx := context.WithValue(c.Request.Context(), requestIDContextKey{}, id)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestIDFromContext returns the request ID stored by the RequestID
// middleware, or an empty string if no ID is present (e.g., a unit test
// that did not pass through the middleware chain).
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey{}).(string)
	return id
}

// isValidUUIDv4 reports whether s is a syntactically valid UUID v4.
// Used to decide whether an incoming X-Request-ID can be trusted as-is
// or must be replaced. Other UUID versions (v1 timestamp-based, v5
// namespace-based) are rejected: the contract is "a v4 or nothing".
func isValidUUIDv4(s string) bool {
	if s == "" {
		return false
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	return parsed.Version() == 4
}
