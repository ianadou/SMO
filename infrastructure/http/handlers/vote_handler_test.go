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

	"github.com/ianadou/smo/application/usecases/vote"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/http/handlers"
)

// --- fakes ---------------------------------------------------------------

type fakeVoteRepo struct {
	mu    sync.Mutex
	votes map[entities.VoteID]*entities.Vote
}

func newFakeVoteRepo() *fakeVoteRepo {
	return &fakeVoteRepo{votes: make(map[entities.VoteID]*entities.Vote)}
}

func (r *fakeVoteRepo) Save(_ context.Context, v *entities.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.votes[v.ID()] = v
	return nil
}

func (r *fakeVoteRepo) FindByID(_ context.Context, id entities.VoteID) (*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.votes[id]
	if !ok {
		return nil, domainerrors.ErrVoteNotFound
	}
	return v, nil
}

func (r *fakeVoteRepo) ListByMatch(_ context.Context, matchID entities.MatchID) ([]*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Vote, 0)
	for _, v := range r.votes {
		if v.MatchID() == matchID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (r *fakeVoteRepo) ListByVoter(_ context.Context, voterID entities.PlayerID) ([]*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Vote, 0)
	for _, v := range r.votes {
		if v.VoterID() == voterID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (r *fakeVoteRepo) Delete(_ context.Context, id entities.VoteID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.votes, id)
	return nil
}

// Minimal MatchRepository with only FindByID exercised.
type voteTestMatchRepo struct {
	match *entities.Match
}

func (r *voteTestMatchRepo) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	if r.match == nil || r.match.ID() != id {
		return nil, domainerrors.ErrMatchNotFound
	}
	return r.match, nil
}

func (r *voteTestMatchRepo) Save(context.Context, *entities.Match) error { panic("unused") }
func (r *voteTestMatchRepo) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	panic("unused")
}
func (r *voteTestMatchRepo) UpdateStatus(context.Context, *entities.Match) error { panic("unused") }
func (r *voteTestMatchRepo) Delete(context.Context, entities.MatchID) error      { panic("unused") }

type voteIDGen struct{ id string }

func (g *voteIDGen) Generate() string { return g.id }

type voteClock struct{ now time.Time }

func (c *voteClock) Now() time.Time { return c.now }

// --- helpers -------------------------------------------------------------

func buildVoteTestRouter(t *testing.T, matchStatus entities.MatchStatus) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	voteRepo := newFakeVoteRepo()
	match, _ := entities.NewMatch("test-match", "g-1", "Test", "V",
		time.Now().Add(24*time.Hour), matchStatus, time.Now())
	matchRepo := &voteTestMatchRepo{match: match}

	handler := handlers.NewVoteHandler(
		vote.NewCastVoteUseCase(voteRepo, matchRepo, &voteIDGen{id: "v-gen"}, &voteClock{now: time.Now()}),
		vote.NewGetVoteUseCase(voteRepo),
		vote.NewListVotesByMatchUseCase(voteRepo),
	)
	router := gin.New()
	api := router.Group("/api")
	handler.Register(api)
	return router
}

// --- tests ---------------------------------------------------------------

func TestVoteHandler_Cast_Returns201_WhenMatchCompleted(t *testing.T) {
	router := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	body := `{"match_id":"test-match","voter_id":"p-1","voted_id":"p-2","score":4}`
	req := httptest.NewRequest(http.MethodPost, "/api/votes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["score"].(float64) != 4 {
		t.Errorf("expected score 4, got %v", resp["score"])
	}
}

func TestVoteHandler_Cast_Returns409_WhenMatchNotCompleted(t *testing.T) {
	router := buildVoteTestRouter(t, entities.MatchStatusDraft)

	body := `{"match_id":"test-match","voter_id":"p-1","voted_id":"p-2","score":4}`
	req := httptest.NewRequest(http.MethodPost, "/api/votes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestVoteHandler_Cast_Returns400_OnSelfVote(t *testing.T) {
	router := buildVoteTestRouter(t, entities.MatchStatusCompleted)

	body := `{"match_id":"test-match","voter_id":"p-1","voted_id":"p-1","score":4}`
	req := httptest.NewRequest(http.MethodPost, "/api/votes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}
