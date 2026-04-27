package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/http/middlewares"
)

type stubSigner struct {
	expectedToken string
	organizerID   entities.OrganizerID
}

func (s *stubSigner) Sign(_ entities.OrganizerID) (string, error) { return s.expectedToken, nil }
func (s *stubSigner) Verify(token string) (entities.OrganizerID, error) {
	if token != s.expectedToken {
		return "", domainerrors.ErrInvalidToken
	}
	return s.organizerID, nil
}

func newRouterWithJWTAuth(signer *stubSigner) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middlewares.JWTAuth(signer))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"organizer_id": string(middlewares.OrganizerIDFromContext(c.Request.Context())),
		})
	})
	return router
}

func TestJWTAuth_Returns401_WhenAuthorizationHeaderMissing(t *testing.T) {
	t.Parallel()
	router := newRouterWithJWTAuth(&stubSigner{expectedToken: "tok", organizerID: "org-1"})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestJWTAuth_Returns401_WhenHeaderMissingBearerPrefix(t *testing.T) {
	t.Parallel()
	router := newRouterWithJWTAuth(&stubSigner{expectedToken: "tok", organizerID: "org-1"})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "tok") // no "Bearer "
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestJWTAuth_Returns401_WhenTokenIsInvalid(t *testing.T) {
	t.Parallel()
	router := newRouterWithJWTAuth(&stubSigner{expectedToken: "valid-tok", organizerID: "org-1"})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestJWTAuth_Returns200_AndExposesOrganizerIDInContext_WhenTokenIsValid(t *testing.T) {
	t.Parallel()
	router := newRouterWithJWTAuth(&stubSigner{expectedToken: "valid-tok", organizerID: "org-42"})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-tok")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	if !contains(rec.Body.String(), "org-42") {
		t.Errorf("expected response to expose organizer_id 'org-42', got %s", rec.Body.String())
	}
}

func contains(body, substr string) bool {
	for i := 0; i+len(substr) <= len(body); i++ {
		if body[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
