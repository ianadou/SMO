package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// allowedHeaders are the request headers the SMO frontend sends.
// Listed explicitly rather than using "*" so a future header is a
// conscious decision, not a silent broaden of the surface.
const allowedHeaders = "Authorization,Content-Type,X-Request-ID"

// allowedMethods covers every verb the API uses today. Kept tight on
// purpose — adding a new verb to the API is a conscious config bump.
const allowedMethods = "GET,POST,PATCH,DELETE,OPTIONS"

// preflightMaxAgeSeconds tells the browser how long it can cache the
// preflight response before re-issuing OPTIONS. Twelve hours is the
// upper bound Chrome respects.
const preflightMaxAgeSeconds = "43200"

// CORS returns a Gin middleware that handles browser cross-origin
// requests for the configured allowed origins. Requests from any
// other origin pass through without CORS headers, so curl and other
// non-browser clients are unaffected.
//
// allowedOrigins is the exact set of origins (scheme + host + port)
// the browser is allowed to send credentials from. The middleware
// does not allow "*" because credentialed requests (Authorization
// header) require an explicit origin per the CORS spec.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		if _, ok := allowed[origin]; !ok {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Max-Age", preflightMaxAgeSeconds)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
