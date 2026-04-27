package match

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ranking"
)

// --- minimal fakes for FinalizeMatchUseCase --------------------------------

type fakeVoteRepoForFinalize struct {
	mu    sync.Mutex
	votes map[entities.MatchID][]*entities.Vote
}

func newFakeVoteRepoForFinalize() *fakeVoteRepoForFinalize {
	return &fakeVoteRepoForFinalize{votes: make(map[entities.MatchID][]*entities.Vote)}
}

func (r *fakeVoteRepoForFinalize) Save(_ context.Context, v *entities.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.votes[v.MatchID()] = append(r.votes[v.MatchID()], v)
	return nil
}

func (r *fakeVoteRepoForFinalize) FindByID(context.Context, entities.VoteID) (*entities.Vote, error) {
	panic("not used")
}

func (r *fakeVoteRepoForFinalize) ListByMatch(_ context.Context, id entities.MatchID) ([]*entities.Vote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*entities.Vote(nil), r.votes[id]...), nil
}

func (r *fakeVoteRepoForFinalize) ListByVoter(context.Context, entities.PlayerID) ([]*entities.Vote, error) {
	panic("not used")
}

func (r *fakeVoteRepoForFinalize) Delete(context.Context, entities.VoteID) error {
	panic("not used")
}

type fakePlayerRepoForFinalize struct {
	mu      sync.Mutex
	players map[entities.PlayerID]*entities.Player
}

func newFakePlayerRepoForFinalize() *fakePlayerRepoForFinalize {
	return &fakePlayerRepoForFinalize{players: make(map[entities.PlayerID]*entities.Player)}
}

func (r *fakePlayerRepoForFinalize) Save(_ context.Context, p *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[p.ID()] = p
	return nil
}

func (r *fakePlayerRepoForFinalize) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *fakePlayerRepoForFinalize) ListByGroup(context.Context, entities.GroupID) ([]*entities.Player, error) {
	panic("not used")
}

func (r *fakePlayerRepoForFinalize) UpdateRanking(_ context.Context, p *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[p.ID()] = p
	return nil
}

func (r *fakePlayerRepoForFinalize) Delete(context.Context, entities.PlayerID) error {
	panic("not used")
}

// --- helper -----------------------------------------------------------------

func newFinalizeUseCase(t *testing.T) (*FinalizeMatchUseCase, *fakeMatchRepository, *fakeVoteRepoForFinalize, *fakePlayerRepoForFinalize) {
	t.Helper()
	matchRepo := newFakeMatchRepository()
	voteRepo := newFakeVoteRepoForFinalize()
	playerRepo := newFakePlayerRepoForFinalize()
	calculator, err := ranking.NewCalculator(ranking.DefaultLearningRate())
	if err != nil {
		t.Fatalf("test setup: build calculator: %v", err)
	}
	uc := NewFinalizeMatchUseCase(matchRepo, voteRepo, playerRepo, calculator)
	return uc, matchRepo, voteRepo, playerRepo
}

func seedCompletedMatch(t *testing.T, repo *fakeMatchRepository, id entities.MatchID) *entities.Match {
	t.Helper()
	m, err := entities.NewMatch(id, "g-1", "Test", "Venue", time.Now().Add(time.Hour),
		entities.MatchStatusCompleted, nil, time.Now())
	if err != nil {
		t.Fatalf("seed: NewMatch: %v", err)
	}
	if saveErr := repo.Save(context.Background(), m); saveErr != nil {
		t.Fatalf("seed: Save: %v", saveErr)
	}
	return m
}

// --- tests -----------------------------------------------------------------

func TestFinalizeMatchUseCase_HappyPath_ReturnsMVPAndUpdatedRankingsAndClosesMatch(t *testing.T) {
	t.Parallel()
	uc, matchRepo, voteRepo, playerRepo := newFinalizeUseCase(t)
	ctx := context.Background()

	playerA, _ := entities.NewPlayer("p-a", "g-1", "Alice", 1000)
	playerB, _ := entities.NewPlayer("p-b", "g-1", "Bob", 1000)
	_ = playerRepo.Save(ctx, playerA)
	_ = playerRepo.Save(ctx, playerB)

	seedCompletedMatch(t, matchRepo, "match-1")

	v1, _ := entities.NewVote("v-1", "match-1", "p-a", "p-b", 5, time.Now())
	_ = voteRepo.Save(ctx, v1)

	out, err := uc.Execute(ctx, "match-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.Match.Status() != entities.MatchStatusClosed {
		t.Errorf("expected match closed, got %q", out.Match.Status())
	}
	if out.MVP == nil || *out.MVP != "p-b" {
		t.Errorf("expected MVP p-b, got %v", out.MVP)
	}
	if out.UpdatedRankings["p-b"] <= 1000 {
		t.Errorf("expected p-b ranking to increase, got %d", out.UpdatedRankings["p-b"])
	}
}

func TestFinalizeMatchUseCase_NoVotes_ReturnsNilMVPAndClosesMatch(t *testing.T) {
	t.Parallel()
	uc, matchRepo, _, _ := newFinalizeUseCase(t)

	seedCompletedMatch(t, matchRepo, "match-1")

	out, err := uc.Execute(context.Background(), "match-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.MVP != nil {
		t.Errorf("expected nil MVP when no votes, got %v", out.MVP)
	}
	if out.Match.Status() != entities.MatchStatusClosed {
		t.Errorf("expected match closed, got %q", out.Match.Status())
	}
}

func TestFinalizeMatchUseCase_RejectsTransition_WhenMatchIsNotCompleted(t *testing.T) {
	t.Parallel()
	uc, matchRepo, _, _ := newFinalizeUseCase(t)

	// Match is in Draft status; finalize must be rejected.
	m, _ := entities.NewMatch("match-1", "g-1", "Test", "Venue",
		time.Now().Add(time.Hour), entities.MatchStatusDraft, nil, time.Now())
	_ = matchRepo.Save(context.Background(), m)

	_, err := uc.Execute(context.Background(), "match-1")

	if !errors.Is(err, domainerrors.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestFinalizeMatchUseCase_ReturnsErrMatchNotFound_WhenMatchMissing(t *testing.T) {
	t.Parallel()
	uc, _, _, _ := newFinalizeUseCase(t)

	_, err := uc.Execute(context.Background(), "nonexistent")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}
