package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/invitation"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/http/handlers"
	"github.com/ianadou/smo/infrastructure/token"
)

// --- fakes ---------------------------------------------------------------

type fakeInvRepo struct {
	mu          sync.Mutex
	invitations map[entities.InvitationID]*entities.Invitation
}

func newFakeInvRepo() *fakeInvRepo {
	return &fakeInvRepo{invitations: make(map[entities.InvitationID]*entities.Invitation)}
}

func (r *fakeInvRepo) Save(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvRepo) FindByID(_ context.Context, id entities.InvitationID) (*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.invitations[id]
	if !ok {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *fakeInvRepo) FindByTokenHash(_ context.Context, hash string) (*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inv := range r.invitations {
		if inv.TokenHash() == hash {
			return inv, nil
		}
	}
	return nil, domainerrors.ErrInvitationNotFound
}

func (r *fakeInvRepo) ListByMatch(_ context.Context, matchID entities.MatchID) ([]*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Invitation, 0)
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *fakeInvRepo) MarkAsUsed(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.invitations[inv.ID()]; !ok {
		return domainerrors.ErrInvitationNotFound
	}
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvRepo) Delete(_ context.Context, id entities.InvitationID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.invitations, id)
	return nil
}

type invFixedIDGen struct{ id string }

func (g *invFixedIDGen) Generate() string { return g.id }

type invFixedClock struct{ now time.Time }

func (c *invFixedClock) Now() time.Time { return c.now }

// --- helpers -------------------------------------------------------------

func buildInvitationTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := newFakeInvRepo()
	tokens := token.New() // real token service for realistic hashing
	idGen := &invFixedIDGen{id: "inv-generated"}
	clock := &invFixedClock{now: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)}

	handler := handlers.NewInvitationHandler(
		invitation.NewCreateInvitationUseCase(repo, tokens, idGen, clock),
		invitation.NewGetInvitationUseCase(repo),
		invitation.NewListInvitationsByMatchUseCase(repo),
		invitation.NewAcceptInvitationUseCase(repo, tokens, clock),
	)

	router := gin.New()
	api := router.Group("/api")
	handler.Register(api)
	return router
}

// --- tests ---------------------------------------------------------------

func TestInvitationHandler_Create_Returns201_WithPlainToken(t *testing.T) {
	router := buildInvitationTestRouter(t)

	body := `{"match_id":"match-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/invitations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["plain_token"] == nil || resp["plain_token"] == "" {
		t.Errorf("expected plain_token in response, got %v", resp["plain_token"])
	}
	if resp["id"] != "inv-generated" {
		t.Errorf("expected id 'inv-generated', got %v", resp["id"])
	}
}

func TestInvitationHandler_Create_Returns400_WhenBodyMissingFields(t *testing.T) {
	router := buildInvitationTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/invitations", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestInvitationHandler_Get_Returns404_WhenMissing(t *testing.T) {
	router := buildInvitationTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/invitations/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestInvitationHandler_Accept_FullLifecycle(t *testing.T) {
	// End-to-end: create → accept → verify used_at populated.
	router := buildInvitationTestRouter(t)

	// 1. Create.
	createBody := `{"match_id":"match-1"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/invitations", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	var createResp map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	plainToken, ok := createResp["plain_token"].(string)
	if !ok || plainToken == "" {
		t.Fatalf("no plain_token in create response: %v", createResp)
	}

	// 2. Accept with the plain token.
	acceptBody := `{"token":"` + plainToken + `"}`
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/invitations/accept", bytes.NewBufferString(acceptBody))
	acceptReq.Header.Set("Content-Type", "application/json")
	acceptRec := httptest.NewRecorder()
	router.ServeHTTP(acceptRec, acceptReq)

	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on accept, got %d (body=%s)", acceptRec.Code, acceptRec.Body.String())
	}
	var acceptResp map[string]any
	_ = json.Unmarshal(acceptRec.Body.Bytes(), &acceptResp)
	if acceptResp["used_at"] == nil {
		t.Error("expected used_at to be populated after accept")
	}
}

func TestInvitationHandler_Accept_Returns404_WhenTokenInvalid(t *testing.T) {
	router := buildInvitationTestRouter(t)

	body := `{"token":"definitely-not-a-real-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/invitations/accept", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestInvitationHandler_ListByMatch_Returns200_WithArray(t *testing.T) {
	router := buildInvitationTestRouter(t)

	// Create one invitation first.
	createReq := httptest.NewRequest(http.MethodPost, "/api/invitations",
		bytes.NewBufferString(`{"match_id":"match-42"}`))
	createReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	req := httptest.NewRequest(http.MethodGet, "/api/matches/match-42/invitations", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var list []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Errorf("expected 1 invitation, got %d", len(list))
	}
}
