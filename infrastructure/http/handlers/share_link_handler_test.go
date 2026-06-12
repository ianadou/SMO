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

	"github.com/ianadou/smo/application/usecases/sharelink"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/http/handlers"
	"github.com/ianadou/smo/infrastructure/http/middlewares"
	"github.com/ianadou/smo/infrastructure/token"
)

// --- fakes ---------------------------------------------------------------

// shareFakeLinkRepo stores links in a map. FindActiveByMatchID filters
// on activity at the configured reference time, like the SQL adapter
// filters on now().
type shareFakeLinkRepo struct {
	mu    sync.Mutex
	links map[entities.MatchShareLinkID]*entities.MatchShareLink
	now   time.Time
}

func newShareFakeLinkRepo(now time.Time) *shareFakeLinkRepo {
	return &shareFakeLinkRepo{links: make(map[entities.MatchShareLinkID]*entities.MatchShareLink), now: now}
}

func (r *shareFakeLinkRepo) Create(_ context.Context, link *entities.MatchShareLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ID()] = link
	return nil
}

func (r *shareFakeLinkRepo) FindByTokenHash(_ context.Context, tokenHash string) (*entities.MatchShareLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, link := range r.links {
		if link.TokenHash() == tokenHash {
			return link, nil
		}
	}
	return nil, domainerrors.ErrShareLinkNotFound
}

func (r *shareFakeLinkRepo) FindActiveByMatchID(_ context.Context, matchID entities.MatchID) (*entities.MatchShareLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, link := range r.links {
		if link.MatchID() == matchID && link.IsActive(r.now) {
			return link, nil
		}
	}
	return nil, domainerrors.ErrShareLinkNotFound
}

func (r *shareFakeLinkRepo) Update(_ context.Context, link *entities.MatchShareLink) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.links[link.ID()]; !ok {
		return domainerrors.ErrShareLinkNotFound
	}
	r.links[link.ID()] = link
	return nil
}

// shareFakePlayerRepo stores players in a map so the join flow can both
// resolve existing group players and persist self-added ones.
type shareFakePlayerRepo struct {
	mu      sync.Mutex
	players map[entities.PlayerID]*entities.Player
}

func newShareFakePlayerRepo() *shareFakePlayerRepo {
	return &shareFakePlayerRepo{players: make(map[entities.PlayerID]*entities.Player)}
}

func (r *shareFakePlayerRepo) seedPlayer(t *testing.T, id entities.PlayerID, name string) {
	t.Helper()
	player, err := entities.NewPlayer(id, "g-1", name, 1000)
	if err != nil {
		t.Fatalf("seedPlayer: %v", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[id] = player
}

func (r *shareFakePlayerRepo) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	player, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return player, nil
}

func (r *shareFakePlayerRepo) Save(_ context.Context, player *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[player.ID()] = player
	return nil
}

func (r *shareFakePlayerRepo) UpdateRanking(context.Context, *entities.Player) error { return nil }
func (r *shareFakePlayerRepo) Delete(context.Context, entities.PlayerID) error       { return nil }

func (r *shareFakePlayerRepo) ListByGroup(_ context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*entities.Player, 0, len(r.players))
	for _, player := range r.players {
		if player.GroupID() == groupID {
			out = append(out, player)
		}
	}
	return out, nil
}

// shareStubMatchRepo answers FindByID with a match in group "g-1" whose
// status is configurable, so tests can flip attendance between open and
// locked without driving the whole transition machine.
type shareStubMatchRepo struct{ status entities.MatchStatus }

func (s shareStubMatchRepo) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	return entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          id,
		GroupID:     "g-1",
		Title:       "Match",
		Venue:       "Venue",
		ScheduledAt: now.Add(2 * 24 * time.Hour),
		Status:      s.status,
		CreatedAt:   now,
	})
}
func (shareStubMatchRepo) Save(context.Context, *entities.Match) error         { return nil }
func (shareStubMatchRepo) UpdateStatus(context.Context, *entities.Match) error { return nil }
func (shareStubMatchRepo) Finalize(context.Context, *entities.Match) error     { return nil }
func (shareStubMatchRepo) ReplaceTeams(context.Context, *entities.Match) error { return nil }
func (shareStubMatchRepo) Delete(context.Context, entities.MatchID) error      { return nil }
func (shareStubMatchRepo) CountClosedMatchesTogether(context.Context, entities.GroupID, entities.PlayerID, []entities.PlayerID) (map[entities.PlayerID]int, error) {
	return map[entities.PlayerID]int{}, nil
}

func (shareStubMatchRepo) FindLatestDecidedByGroup(context.Context, entities.GroupID, entities.MatchID) (*entities.Match, error) {
	return nil, domainerrors.ErrMatchNotFound
}

func (shareStubMatchRepo) ListTeamMembersWithPlayers(context.Context, entities.MatchID) ([]entities.MatchTeamMember, error) {
	return nil, nil
}

func (shareStubMatchRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	return nil, nil
}

// --- helpers -------------------------------------------------------------

type shareLinkTestEnv struct {
	router      *gin.Engine
	invitations *fakeInvRepo
	players     *shareFakePlayerRepo
	clockNow    time.Time
}

// buildShareLinkTestRouter wires the share link handler over fakes. The
// organizer ID is injected under the JWTAuth context key when non-empty,
// mirroring what the real middleware does after token verification.
func buildShareLinkTestRouter(t *testing.T, organizerID entities.OrganizerID, matchStatus entities.MatchStatus) *shareLinkTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	clock := &invFixedClock{now: now}
	tokens := token.New()
	idGen := &invFixedIDGen{id: "share-generated"}
	links := newShareFakeLinkRepo(now)
	invitations := newFakeInvRepo()
	players := newShareFakePlayerRepo()
	matchRepo := shareStubMatchRepo{status: matchStatus}

	handler := handlers.NewShareLinkHandler(
		sharelink.NewGenerateMatchShareLinkUseCase(links, matchRepo, invStubGroupRepo{}, tokens, idGen, clock),
		sharelink.NewRevokeMatchShareLinkUseCase(links, matchRepo, invStubGroupRepo{}, clock),
		sharelink.NewGetShareLinkContextUseCase(links, invitations, matchRepo, invStubGroupRepo{}, invStubOrganizerRepo{}, players, tokens, clock),
		sharelink.NewClaimInvitationUseCase(links, invitations, matchRepo, players, tokens, clock),
		sharelink.NewJoinMatchUseCase(links, invitations, matchRepo, players, tokens, idGen, clock),
	)

	router := gin.New()
	public := router.Group("/api/v1")
	protected := router.Group("/api/v1")
	if organizerID != "" {
		protected.Use(func(c *gin.Context) {
			c.Request = c.Request.WithContext(middlewares.WithOrganizerID(c.Request.Context(), organizerID))
			c.Next()
		})
	}
	handler.Register(public, protected)

	return &shareLinkTestEnv{router: router, invitations: invitations, players: players, clockNow: now}
}

// generateShareToken creates a share link through the HTTP API and
// returns its one-time plain token.
func generateShareToken(t *testing.T, env *shareLinkTestEnv) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match-1/share-link", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("generate share link: expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	plainToken, ok := resp["token"].(string)
	if !ok || plainToken == "" {
		t.Fatalf("no token in generate response: %v", resp)
	}
	return plainToken
}

// seedClaimableInvitation stores a pending, unclaimed invitation for the
// player on match-1, plus the player itself in group g-1.
func seedClaimableInvitation(t *testing.T, env *shareLinkTestEnv, playerID entities.PlayerID, playerName string) {
	t.Helper()
	env.players.seedPlayer(t, playerID, playerName)
	inv, err := entities.NewInvitation(
		entities.InvitationID("inv-"+string(playerID)), "match-1", playerID, "seed-hash-"+string(playerID),
		env.clockNow.Add(5*24*time.Hour), entities.InvitationResponsePending, nil, nil, env.clockNow,
	)
	if err != nil {
		t.Fatalf("seedClaimableInvitation: %v", err)
	}
	if saveErr := env.invitations.Save(context.Background(), inv); saveErr != nil {
		t.Fatalf("seedClaimableInvitation: save: %v", saveErr)
	}
}

func postJSON(t *testing.T, router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// --- generate ------------------------------------------------------------

func TestShareLinkHandler_Generate_Returns201_WithPlainTokenAndExpiry(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match-1/share-link", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Errorf("expected plain token in response, got %v", resp["token"])
	}
	if resp["expires_at"] == nil {
		t.Errorf("expected expires_at in response, got %v", resp)
	}
}

func TestShareLinkHandler_Generate_Returns401_WhenOrganizerMissing(t *testing.T) {
	env := buildShareLinkTestRouter(t, "", entities.MatchStatusOpen)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match-1/share-link", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestShareLinkHandler_Generate_Returns404_WhenMatchBelongsToAnotherOrganizer(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-intruder", entities.MatchStatusOpen)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match-1/share-link", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// --- revoke --------------------------------------------------------------

func TestShareLinkHandler_Revoke_Returns204_WhenActiveLinkExists(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	generateShareToken(t, env)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/matches/match-1/share-link", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestShareLinkHandler_Revoke_Returns404_WhenNoActiveLink(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/matches/match-1/share-link", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- context -------------------------------------------------------------

func TestShareLinkHandler_Context_Returns200_WithRosterStates(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	seedClaimableInvitation(t, env, "p-1", "Karim")
	shareToken := generateShareToken(t, env)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/"+shareToken, nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		MatchID       string `json:"match_id"`
		OrganizerName string `json:"organizer_name"`
		MatchStatus   string `json:"match_status"`
		Roster        []struct {
			PlayerID   string `json:"player_id"`
			PlayerName string `json:"player_name"`
			State      string `json:"state"`
		} `json:"roster"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal context response: %v", err)
	}
	if resp.MatchID != "match-1" {
		t.Errorf("expected match_id 'match-1', got %q", resp.MatchID)
	}
	if resp.OrganizerName != "Eddin" {
		t.Errorf("expected organizer 'Eddin', got %q", resp.OrganizerName)
	}
	if resp.MatchStatus != "open" {
		t.Errorf("expected match_status 'open', got %q", resp.MatchStatus)
	}
	if len(resp.Roster) != 1 {
		t.Fatalf("expected 1 roster entry, got %d", len(resp.Roster))
	}
	if resp.Roster[0].PlayerID != "p-1" || resp.Roster[0].PlayerName != "Karim" || resp.Roster[0].State != "claimable" {
		t.Errorf("unexpected roster entry: %+v", resp.Roster[0])
	}
}

func TestShareLinkHandler_Context_Returns404_WhenTokenUnknown(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/unknown-token", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestShareLinkHandler_Context_Returns404_WhenLinkRevoked(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	shareToken := generateShareToken(t, env)

	revokeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/matches/match-1/share-link", nil)
	revokeRec := httptest.NewRecorder()
	env.router.ServeHTTP(revokeRec, revokeReq)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/"+shareToken, nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for revoked link, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] != "share link not found" {
		t.Errorf("revoked link must be indistinguishable from unknown, got %v", resp["error"])
	}
}

// --- claim ---------------------------------------------------------------

func TestShareLinkHandler_Claim_Returns200_WithInvitationToken(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	seedClaimableInvitation(t, env, "p-1", "Karim")
	shareToken := generateShareToken(t, env)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/claim", `{"player_id":"p-1"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["invitation_token"] == nil || resp["invitation_token"] == "" {
		t.Errorf("expected invitation_token in response, got %v", resp)
	}
}

func TestShareLinkHandler_Claim_Returns409_WhenAlreadyClaimed(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	seedClaimableInvitation(t, env, "p-1", "Karim")
	shareToken := generateShareToken(t, env)
	postJSON(t, env.router, "/api/v1/share/"+shareToken+"/claim", `{"player_id":"p-1"}`)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/claim", `{"player_id":"p-1"}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestShareLinkHandler_Claim_Returns423_WhenAttendanceLocked(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusInProgress)
	seedClaimableInvitation(t, env, "p-1", "Karim")
	shareToken := generateShareToken(t, env)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/claim", `{"player_id":"p-1"}`)

	if rec.Code != http.StatusLocked {
		t.Errorf("expected 423, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestShareLinkHandler_Claim_Returns400_WhenPlayerIDMissing(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	shareToken := generateShareToken(t, env)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/claim", `{}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestShareLinkHandler_Claim_Returns404_WhenLinkUnknown(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	seedClaimableInvitation(t, env, "p-1", "Karim")

	rec := postJSON(t, env.router, "/api/v1/share/unknown-token/claim", `{"player_id":"p-1"}`)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- join ----------------------------------------------------------------

func TestShareLinkHandler_Join_Returns201_WithInvitationToken(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	shareToken := generateShareToken(t, env)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/join", `{"player_name":"Nouveau"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["invitation_token"] == nil || resp["invitation_token"] == "" {
		t.Errorf("expected invitation_token in response, got %v", resp)
	}
}

func TestShareLinkHandler_Join_Returns409_WhenNameAlreadyInvited(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	seedClaimableInvitation(t, env, "p-7", "Leila")
	shareToken := generateShareToken(t, env)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/join", `{"player_name":"leila"}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestShareLinkHandler_Join_Returns423_WhenAttendanceLocked(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusInProgress)
	shareToken := generateShareToken(t, env)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/join", `{"player_name":"Nouveau"}`)

	if rec.Code != http.StatusLocked {
		t.Errorf("expected 423, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestShareLinkHandler_Join_Returns400_WhenNameMissing(t *testing.T) {
	env := buildShareLinkTestRouter(t, "org-1", entities.MatchStatusOpen)
	shareToken := generateShareToken(t, env)

	rec := postJSON(t, env.router, "/api/v1/share/"+shareToken+"/join", `{}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
