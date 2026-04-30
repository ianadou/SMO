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
	"github.com/ianadou/smo/application/usecases/match"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ranking"
	"github.com/ianadou/smo/infrastructure/http/handlers"
)

// --- fakes ---------------------------------------------------------------

type fakeMatchRepo struct {
	mu      sync.Mutex
	matches map[entities.MatchID]*entities.Match
}

func newFakeMatchRepo() *fakeMatchRepo {
	return &fakeMatchRepo{matches: make(map[entities.MatchID]*entities.Match)}
}

func (r *fakeMatchRepo) Save(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[m.ID()] = m
	return nil
}

func (r *fakeMatchRepo) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *fakeMatchRepo) ListByGroup(_ context.Context, groupID entities.GroupID) ([]*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Match, 0)
	for _, m := range r.matches {
		if m.GroupID() == groupID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *fakeMatchRepo) UpdateStatus(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.matches[m.ID()]; !ok {
		return domainerrors.ErrMatchNotFound
	}
	r.matches[m.ID()] = m
	return nil
}

func (r *fakeMatchRepo) Finalize(_ context.Context, m *entities.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.matches[m.ID()]; !ok {
		return domainerrors.ErrMatchNotFound
	}
	r.matches[m.ID()] = m
	return nil
}

func (r *fakeMatchRepo) Delete(_ context.Context, id entities.MatchID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.matches, id)
	return nil
}

// finalizeFakeVoteRepo / finalizeFakePlayerRepo are minimal stubs, only used by the
// FinalizeMatchUseCase wired through the handler in TestMatchHandler_Finalize.

type finalizeFakeVoteRepo struct {
	votes map[entities.MatchID][]*entities.Vote
}

func newFinalizeFakeVoteRepo() *finalizeFakeVoteRepo {
	return &finalizeFakeVoteRepo{votes: make(map[entities.MatchID][]*entities.Vote)}
}

func (r *finalizeFakeVoteRepo) Save(context.Context, *entities.Vote) error {
	panic("not used in handler tests")
}

func (r *finalizeFakeVoteRepo) FindByID(context.Context, entities.VoteID) (*entities.Vote, error) {
	panic("not used in handler tests")
}

func (r *finalizeFakeVoteRepo) ListByMatch(_ context.Context, matchID entities.MatchID) ([]*entities.Vote, error) {
	return r.votes[matchID], nil
}

func (r *finalizeFakeVoteRepo) ListByVoter(context.Context, entities.PlayerID) ([]*entities.Vote, error) {
	panic("not used in handler tests")
}

func (r *finalizeFakeVoteRepo) Delete(context.Context, entities.VoteID) error {
	panic("not used in handler tests")
}

type finalizeFakePlayerRepo struct {
	players map[entities.PlayerID]*entities.Player
}

func newFinalizeFakePlayerRepo() *finalizeFakePlayerRepo {
	return &finalizeFakePlayerRepo{players: make(map[entities.PlayerID]*entities.Player)}
}

func (r *finalizeFakePlayerRepo) Save(_ context.Context, p *entities.Player) error {
	r.players[p.ID()] = p
	return nil
}

func (r *finalizeFakePlayerRepo) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *finalizeFakePlayerRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Player, error) {
	panic("not used in handler tests")
}

func (r *finalizeFakePlayerRepo) UpdateRanking(_ context.Context, p *entities.Player) error {
	r.players[p.ID()] = p
	return nil
}

func (r *finalizeFakePlayerRepo) Delete(context.Context, entities.PlayerID) error {
	panic("not used in handler tests")
}

type fixedIDGen struct{ id string }

func (g *fixedIDGen) Generate() string { return g.id }

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

// --- helpers -------------------------------------------------------------

// testRouter is the bag of wired pieces returned by buildTestRouter so
// tests can pre-seed matches, votes, and players before calling the API.
type testRouter struct {
	router     *gin.Engine
	matchRepo  *fakeMatchRepo
	voteRepo   *finalizeFakeVoteRepo
	playerRepo *finalizeFakePlayerRepo
}

// buildTestRouter builds a router pre-wired with the Match handler and
// the fake repositories needed to exercise the full lifecycle including
// finalize.
func buildTestRouter(t *testing.T) *testRouter {
	t.Helper()
	gin.SetMode(gin.TestMode)

	matchRepo := newFakeMatchRepo()
	voteRepo := newFinalizeFakeVoteRepo()
	playerRepo := newFinalizeFakePlayerRepo()
	idGen := &fixedIDGen{id: "generated-id"}
	clock := &fixedClock{now: time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)}

	calculator, err := ranking.NewCalculator(ranking.DefaultLearningRate())
	if err != nil {
		t.Fatalf("test setup: build calculator: %v", err)
	}

	handler := handlers.NewMatchHandler(
		match.NewCreateMatchUseCase(matchRepo, idGen, clock),
		match.NewGetMatchUseCase(matchRepo),
		match.NewListMatchesByGroupUseCase(matchRepo),
		match.NewOpenMatchUseCase(matchRepo),
		match.NewMarkTeamsReadyUseCase(matchRepo, noopPublisher{}, clock),
		match.NewStartMatchUseCase(matchRepo),
		match.NewCompleteMatchUseCase(matchRepo),
		match.NewFinalizeMatchUseCase(matchRepo, voteRepo, playerRepo, calculator),
		invitation.NewListMatchParticipantsUseCase(newFakeInvRepo()),
	)

	router := gin.New()
	api := router.Group("/api/v1")
	handler.Register(api, api)
	return &testRouter{router: router, matchRepo: matchRepo, voteRepo: voteRepo, playerRepo: playerRepo}
}

func seedMatch(t *testing.T, repo *fakeMatchRepo) *entities.Match {
	t.Helper()
	m, err := entities.NewMatch(
		"match-1", "group-1", "Test", "Venue",
		time.Now().Add(24*time.Hour), time.Now(),
	)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = repo.Save(context.Background(), m)
	return m
}

// --- tests ---------------------------------------------------------------

func TestMatchHandler_Create_Returns201_WithMatchResponse(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router

	body := `{"group_id":"group-1","title":"Friday","venue":"Stadium","scheduled_at":"2026-05-01T18:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["id"] != "generated-id" {
		t.Errorf("expected id 'generated-id', got %v", resp["id"])
	}
	if resp["status"] != "draft" {
		t.Errorf("expected status 'draft', got %v", resp["status"])
	}
}

func TestMatchHandler_Create_Returns400_WhenBodyIsMissingFields(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", bytes.NewBufferString(`{"title":"Only title"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMatchHandler_Get_Returns200_WhenMatchExists(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router
	seedMatch(t, tr.matchRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/match-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMatchHandler_Get_Returns404_WhenMatchDoesNotExist(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestMatchHandler_ListByGroup_Returns200_WithArray(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router
	seedMatch(t, tr.matchRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/group-1/matches", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var list []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Errorf("expected 1 match, got %d", len(list))
	}
}

func TestMatchHandler_Open_Returns200_AndTransitionsStatus(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router
	seedMatch(t, tr.matchRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match-1/open", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "open" {
		t.Errorf("expected status 'open', got %v", resp["status"])
	}
}

func TestMatchHandler_Start_Returns409_WhenTransitionIsInvalid(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router
	// Start requires TeamsReady; attempting from Draft must return 409 Conflict.
	seedMatch(t, tr.matchRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match-1/start", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestMatchHandler_AllTransitions_ReturnsOK_InOrder(t *testing.T) {
	// Walks the lifecycle up to Completed via the pure-status transitions.
	// The Completed → Closed step is exercised by Finalize, which has its
	// own dedicated test below because it requires votes and players.
	tr := buildTestRouter(t)
	router := tr.router
	seedMatch(t, tr.matchRepo)

	steps := []struct {
		path   string
		status string
	}{
		{"/api/v1/matches/match-1/open", "open"},
		{"/api/v1/matches/match-1/teams-ready", "teams_ready"},
		{"/api/v1/matches/match-1/start", "in_progress"},
		{"/api/v1/matches/match-1/complete", "completed"},
	}

	for _, step := range steps {
		req := httptest.NewRequest(http.MethodPost, step.path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("step %s: expected 200, got %d (body=%s)", step.path, rec.Code, rec.Body.String())
		}
		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["status"] != step.status {
			t.Errorf("step %s: expected status %q, got %v", step.path, step.status, resp["status"])
		}
	}
}

func TestMatchHandler_Finalize_Returns200_WithMVPAndUpdatedRankings(t *testing.T) {
	tr := buildTestRouter(t)
	router := tr.router
	ctx := context.Background()

	playerA, _ := entities.NewPlayer("p-a", "group-1", "Alice", 1000)
	playerB, _ := entities.NewPlayer("p-b", "group-1", "Bob", 1000)
	playerC, _ := entities.NewPlayer("p-c", "group-1", "Carol", 1000)
	_ = tr.playerRepo.Save(ctx, playerA)
	_ = tr.playerRepo.Save(ctx, playerB)
	_ = tr.playerRepo.Save(ctx, playerC)

	completedMatch, _ := entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          "match-1",
		GroupID:     "group-1",
		Title:       "Test",
		Venue:       "Venue",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      entities.MatchStatusCompleted,
		CreatedAt:   time.Now(),
	})
	_ = tr.matchRepo.Save(ctx, completedMatch)

	v1, _ := entities.NewVote("v-1", "match-1", "p-a", "p-b", 5, time.Now())
	v2, _ := entities.NewVote("v-2", "match-1", "p-c", "p-b", 5, time.Now())
	v3, _ := entities.NewVote("v-3", "match-1", "p-a", "p-c", 3, time.Now())
	tr.voteRepo.votes["match-1"] = []*entities.Vote{v1, v2, v3}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/match-1/finalize", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	matchSection, ok := resp["match"].(map[string]any)
	if !ok {
		t.Fatalf("response missing 'match' section: %v", resp)
	}
	if matchSection["status"] != "closed" {
		t.Errorf("expected match status 'closed', got %v", matchSection["status"])
	}
	if resp["mvp_player_id"] != "p-b" {
		t.Errorf("expected MVP 'p-b', got %v", resp["mvp_player_id"])
	}
	rankings, ok := resp["updated_rankings"].(map[string]any)
	if !ok {
		t.Fatalf("response missing 'updated_rankings': %v", resp)
	}
	if _, hasB := rankings["p-b"]; !hasB {
		t.Errorf("expected p-b to appear in updated_rankings, got %v", rankings)
	}
}

func TestMatchHandler_Participants_Returns200_WithEmptyArray_WhenNoOneConfirmed(t *testing.T) {
	tr := buildTestRouter(t)
	seedMatch(t, tr.matchRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/match-1/participants", nil)
	rec := httptest.NewRecorder()
	tr.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp []any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp == nil {
		t.Errorf("expected JSON array (possibly empty), got null")
	}
}
