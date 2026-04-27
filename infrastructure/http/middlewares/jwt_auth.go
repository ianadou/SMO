package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// organizerIDContextKey is the context.Context key under which the
// authenticated organizer ID is stored. Typed key prevents collisions
// with other packages.
type organizerIDContextKey struct{}

// JWTAuth returns a Gin middleware that requires a valid Bearer token
// on the Authorization header. It extracts the organizer ID from the
// token, exposes it via context.Context, and aborts with 401 if the
// token is missing or invalid.
//
// This middleware is created in PR #33 but only applied to the
// mutation routes in PR #34. Public routes (auth, accept invitation,
// cast vote) stay unprotected.
func JWTAuth(signer ports.JWTSigner) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authorizationHeader)
		if !strings.HasPrefix(header, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				httperrors.ErrorResponse{Error: "missing or malformed Authorization header"})
			return
		}

		token := strings.TrimPrefix(header, bearerPrefix)
		organizerID, err := signer.Verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				httperrors.ErrorResponse{Error: "invalid token"})
			return
		}

		ctx := context.WithValue(c.Request.Context(), organizerIDContextKey{}, organizerID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("organizer_id", string(organizerID))

		c.Next()
	}
}

// OrganizerIDFromContext returns the authenticated organizer ID stored
// by the JWTAuth middleware, or an empty string if no ID is present
// (the request did not pass through JWTAuth, or the type was wrong).
func OrganizerIDFromContext(ctx context.Context) entities.OrganizerID {
	id, _ := ctx.Value(organizerIDContextKey{}).(entities.OrganizerID)
	return id
}
