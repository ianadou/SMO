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

func (r *fakeInvRepo) CountConfirmedByMatch(_ context.Context, matchID entities.MatchID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID && inv.IsUsed() {
			count++
		}
	}
	return count, nil
}

func (r *fakeInvRepo) ListConfirmedParticipants(_ context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]entities.MatchParticipant, 0)
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID && inv.IsUsed() {
			out = append(out, entities.MatchParticipant{
				PlayerID:    inv.PlayerID(),
				PlayerName:  "Fake " + string(inv.PlayerID()),
				ConfirmedAt: *inv.UsedAt(),
			})
		}
	}
	return out, nil
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

// invStubMatchRepo answers FindByID with a generated Match owned by
// group "g-1", whatever the requested id. The handler tests do not
// exercise the cross-group rejection branch (covered by the use case
// unit tests), so a permissive stub keeps the HTTP-level tests focused
// on routing and serialization.
type invStubMatchRepo struct{}

func (invStubMatchRepo) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	scheduled := now.Add(2 * 24 * time.Hour)
	return entities.NewMatch(id, "g-1", "Match", "Venue", scheduled, now)
}
func (invStubMatchRepo) Save(context.Context, *entities.Match) error         { return nil }
func (invStubMatchRepo) UpdateStatus(context.Context, *entities.Match) error { return nil }
func (invStubMatchRepo) Finalize(context.Context, *entities.Match) error     { return nil }
func (invStubMatchRepo) Delete(context.Context, entities.MatchID) error      { return nil }
func (invStubMatchRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	return nil, nil
}

// invStubPlayerRepo answers FindByID with a Player in group "g-1" so
// the cross-group check inside CreateInvitation always succeeds at
// HTTP level.
type invStubPlayerRepo struct{}

func (invStubPlayerRepo) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	return entities.NewPlayer(id, "g-1", "Stub Player", 1000)
}
func (invStubPlayerRepo) Save(context.Context, *entities.Player) error          { return nil }
func (invStubPlayerRepo) UpdateRanking(context.Context, *entities.Player) error { return nil }
func (invStubPlayerRepo) Delete(context.Context, entities.PlayerID) error       { return nil }
func (invStubPlayerRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Player, error) {
	return nil, nil
}

// --- helpers -------------------------------------------------------------

func buildInvitationTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := newFakeInvRepo()
	tokens := token.New() // real token service for realistic hashing
	idGen := &invFixedIDGen{id: "inv-generated"}
	clock := &invFixedClock{now: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)}

	handler := handlers.NewInvitationHandler(
		invitation.NewCreateInvitationUseCase(repo, invStubMatchRepo{}, invStubPlayerRepo{}, tokens, idGen, clock),
		invitation.NewGetInvitationUseCase(repo),
		invitation.NewListInvitationsByMatchUseCase(repo),
		invitation.NewAcceptInvitationUseCase(repo, tokens, clock),
	)

	router := gin.New()
	api := router.Group("/api/v1")
	handler.Register(api, api)
	return router
}

// --- tests ---------------------------------------------------------------

func TestInvitationHandler_Create_Returns201_WithPlainToken(t *testing.T) {
	router := buildInvitationTestRouter(t)

	body := `{"match_id":"match-1","player_id":"player-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invitations", bytes.NewBufferString(body))
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/invitations", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestInvitationHandler_Get_Returns404_WhenMissing(t *testing.T) {
	router := buildInvitationTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/invitations/nonexistent", nil)
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
	createBody := `{"match_id":"match-1","player_id":"player-1"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/invitations", bytes.NewBufferString(createBody))
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
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/v1/invitations/accept", bytes.NewBufferString(acceptBody))
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invitations/accept", bytes.NewBufferString(body))
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
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/invitations",
		bytes.NewBufferString(`{"match_id":"match-42","player_id":"player-42"}`))
	createReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/match-42/invitations", nil)
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
