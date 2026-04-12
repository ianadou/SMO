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

	"github.com/ianadou/smo/application/usecases/match"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
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

func (r *fakeMatchRepo) Delete(_ context.Context, id entities.MatchID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.matches, id)
	return nil
}

type fixedIDGen struct{ id string }

func (g *fixedIDGen) Generate() string { return g.id }

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

// --- helpers -------------------------------------------------------------

// buildTestRouter builds a router pre-wired with the Match handler and
// the given repository. It returns the router and the repo so tests can
// pre-seed matches directly before calling the API.
func buildTestRouter(t *testing.T) (*gin.Engine, *fakeMatchRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := newFakeMatchRepo()
	idGen := &fixedIDGen{id: "generated-id"}
	clock := &fixedClock{now: time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)}

	handler := handlers.NewMatchHandler(
		match.NewCreateMatchUseCase(repo, idGen, clock),
		match.NewGetMatchUseCase(repo),
		match.NewListMatchesByGroupUseCase(repo),
		match.NewOpenMatchUseCase(repo),
		match.NewMarkTeamsReadyUseCase(repo),
		match.NewStartMatchUseCase(repo),
		match.NewCompleteMatchUseCase(repo),
		match.NewCloseMatchUseCase(repo),
	)

	router := gin.New()
	api := router.Group("/api")
	handler.Register(api)
	return router, repo
}

func seedMatch(t *testing.T, repo *fakeMatchRepo) *entities.Match {
	t.Helper()
	m, err := entities.NewMatch(
		"match-1", "group-1", "Test", "Venue",
		time.Now().Add(24*time.Hour), entities.MatchStatusDraft, time.Now(),
	)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = repo.Save(context.Background(), m)
	return m
}

// --- tests ---------------------------------------------------------------

func TestMatchHandler_Create_Returns201_WithMatchResponse(t *testing.T) {
	router, _ := buildTestRouter(t)

	body := `{"group_id":"group-1","title":"Friday","venue":"Stadium","scheduled_at":"2026-05-01T18:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/matches", bytes.NewBufferString(body))
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
	router, _ := buildTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/matches", bytes.NewBufferString(`{"title":"Only title"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMatchHandler_Get_Returns200_WhenMatchExists(t *testing.T) {
	router, repo := buildTestRouter(t)
	seedMatch(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/matches/match-1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMatchHandler_Get_Returns404_WhenMatchDoesNotExist(t *testing.T) {
	router, _ := buildTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/matches/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestMatchHandler_ListByGroup_Returns200_WithArray(t *testing.T) {
	router, repo := buildTestRouter(t)
	seedMatch(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/api/groups/group-1/matches", nil)
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
	router, repo := buildTestRouter(t)
	seedMatch(t, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/matches/match-1/open", nil)
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
	router, repo := buildTestRouter(t)
	// Start requires TeamsReady; attempting from Draft must return 409 Conflict.
	seedMatch(t, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/matches/match-1/start", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestMatchHandler_AllTransitions_ReturnsOK_InOrder(t *testing.T) {
	// Walks the full lifecycle Draft → Open → TeamsReady → InProgress → Completed → Closed
	// via the HTTP endpoints, checking that each step returns 200 and the expected status.
	router, repo := buildTestRouter(t)
	seedMatch(t, repo)

	steps := []struct {
		path   string
		status string
	}{
		{"/api/matches/match-1/open", "open"},
		{"/api/matches/match-1/teams-ready", "teams_ready"},
		{"/api/matches/match-1/start", "in_progress"},
		{"/api/matches/match-1/complete", "completed"},
		{"/api/matches/match-1/close", "closed"},
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
