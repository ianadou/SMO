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
		if inv.MatchID() == matchID && inv.IsConfirmed() {
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
		if inv.MatchID() == matchID && inv.IsConfirmed() {
			out = append(out, entities.MatchParticipant{
				PlayerID:    inv.PlayerID(),
				PlayerName:  "Fake " + string(inv.PlayerID()),
				ConfirmedAt: *inv.RespondedAt(),
			})
		}
	}
	return out, nil
}

func (r *fakeInvRepo) RespondWithCapacityGuard(_ context.Context, inv *entities.Invitation, _ int) error {
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

// invStubGroupRepo answers FindByID with group "g-1" owned by organizer
// "org-1", matching the group invStubMatchRepo puts every match in.
type invStubGroupRepo struct{}

func (invStubGroupRepo) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	return entities.NewGroup(id, "Les Bras Cassés", "org-1", "", now)
}
func (invStubGroupRepo) Save(context.Context, *entities.Group) error    { return nil }
func (invStubGroupRepo) Update(context.Context, *entities.Group) error  { return nil }
func (invStubGroupRepo) Delete(context.Context, entities.GroupID) error { return nil }
func (invStubGroupRepo) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	return nil, nil
}

// invStubOrganizerRepo answers FindByID with a fixed display name so the
// context endpoint has an organizer to surface.
type invStubOrganizerRepo struct{}

func (invStubOrganizerRepo) FindByID(_ context.Context, id entities.OrganizerID) (*entities.Organizer, error) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	return entities.NewOrganizer(id, "organizer@example.com", "Eddin", "hash", now)
}
func (invStubOrganizerRepo) Save(context.Context, *entities.Organizer) error { return nil }
func (invStubOrganizerRepo) FindByEmail(context.Context, string) (*entities.Organizer, error) {
	return nil, domainerrors.ErrOrganizerNotFound
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
		invitation.NewGetInvitationContextUseCase(repo, invStubMatchRepo{}, invStubGroupRepo{}, invStubOrganizerRepo{}, tokens, clock),
		invitation.NewListInvitationsByMatchUseCase(repo),
		invitation.NewRespondToInvitationUseCase(repo, invStubMatchRepo{}, tokens, clock),
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

// createInvitationAndToken creates an invitation through the HTTP API
// and returns its one-time plain token.
func createInvitationAndToken(t *testing.T, router *gin.Engine) string {
	t.Helper()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/invitations",
		bytes.NewBufferString(`{"match_id":"match-1","player_id":"player-1"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	var createResp map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &createResp)
	plainToken, ok := createResp["plain_token"].(string)
	if !ok || plainToken == "" {
		t.Fatalf("no plain_token in create response: %v", createResp)
	}
	return plainToken
}

func respond(t *testing.T, router *gin.Engine, token, answer string) *httptest.ResponseRecorder {
	t.Helper()
	body := `{"token":"` + token + `","answer":"` + answer + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invitations/respond", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestInvitationHandler_Respond_ConfirmsAttendance_WhenAnswerIsYes(t *testing.T) {
	router := buildInvitationTestRouter(t)
	token := createInvitationAndToken(t, router)

	rec := respond(t, router, token, "yes")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on respond, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["response"] != "yes" {
		t.Errorf("expected response 'yes', got %v", resp["response"])
	}
	if resp["responded_at"] == nil {
		t.Error("expected responded_at to be populated after responding")
	}
}

func TestInvitationHandler_Respond_AllowsChangingTheAnswer(t *testing.T) {
	router := buildInvitationTestRouter(t)
	token := createInvitationAndToken(t, router)

	if rec := respond(t, router, token, "yes"); rec.Code != http.StatusOK {
		t.Fatalf("first respond (yes) failed: %d (%s)", rec.Code, rec.Body.String())
	}
	rec := respond(t, router, token, "no")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on changed answer, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["response"] != "no" {
		t.Errorf("expected response 'no' after change, got %v", resp["response"])
	}
}

func TestInvitationHandler_Respond_Returns400_WhenAnswerInvalid(t *testing.T) {
	router := buildInvitationTestRouter(t)
	token := createInvitationAndToken(t, router)

	rec := respond(t, router, token, "maybe")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid answer, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestInvitationHandler_Respond_Returns404_WhenTokenInvalid(t *testing.T) {
	router := buildInvitationTestRouter(t)

	rec := respond(t, router, "definitely-not-a-real-token", "yes")

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

func requestContext(t *testing.T, router *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invitations/context", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestInvitationHandler_Context_Returns200_WithAssembledContext(t *testing.T) {
	router := buildInvitationTestRouter(t)
	token := createInvitationAndToken(t, router)

	rec := requestContext(t, router, `{"token":"`+token+`"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on context, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["organizer_name"] != "Eddin" {
		t.Errorf("organizer_name = %v, want Eddin", resp["organizer_name"])
	}
	if resp["group_name"] != "Les Bras Cassés" {
		t.Errorf("group_name = %v, want Les Bras Cassés", resp["group_name"])
	}
	if resp["capacity"] != "10 (5v5)" {
		t.Errorf("capacity = %v, want 10 (5v5)", resp["capacity"])
	}
	if resp["state"] != "respondable" {
		t.Errorf("state = %v, want respondable", resp["state"])
	}
	if resp["response"] != "pending" {
		t.Errorf("response = %v, want pending", resp["response"])
	}
}

func TestInvitationHandler_Context_Returns404_WhenTokenUnknown(t *testing.T) {
	router := buildInvitationTestRouter(t)

	rec := requestContext(t, router, `{"token":"does-not-exist"}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown token, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestInvitationHandler_Context_Returns400_WhenTokenMissing(t *testing.T) {
	router := buildInvitationTestRouter(t)

	rec := requestContext(t, router, `{}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing token, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}
