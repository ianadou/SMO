package vote

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func seedCompletedMatch(t *testing.T, repo *fakeMatchRepository) {
	t.Helper()
	m, err := entities.NewMatch("test-match", "g-1", "Test", "V",
		time.Now().Add(24*time.Hour), entities.MatchStatusCompleted, nil, time.Now())
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo.addMatch(m)
}

func seedDraftMatch(t *testing.T, repo *fakeMatchRepository) {
	t.Helper()
	m, err := entities.NewMatch("test-match", "g-1", "Test", "V",
		time.Now().Add(24*time.Hour), entities.MatchStatusDraft, nil, time.Now())
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo.addMatch(m)
}

func TestCastVoteUseCase_Execute_ReturnsVote_WhenMatchIsCompleted(t *testing.T) {
	t.Parallel()
	voteRepo := newFakeVoteRepository()
	matchRepo := newFakeMatchRepository()
	seedCompletedMatch(t, matchRepo)
	uc := NewCastVoteUseCase(voteRepo, matchRepo, newFakeIDGenerator("v-1"),
		newFakeClock(time.Now()))

	vote, err := uc.Execute(context.Background(), CastVoteInput{
		MatchID: "test-match", VoterID: "p-1", VotedID: "p-2", Score: 4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vote.Score() != 4 {
		t.Errorf("expected 4, got %d", vote.Score())
	}
}

func TestCastVoteUseCase_Execute_ReturnsErrMatchNotCompleted_WhenMatchIsDraft(t *testing.T) {
	t.Parallel()
	voteRepo := newFakeVoteRepository()
	matchRepo := newFakeMatchRepository()
	seedDraftMatch(t, matchRepo)
	uc := NewCastVoteUseCase(voteRepo, matchRepo, newFakeIDGenerator("v-1"),
		newFakeClock(time.Now()))

	_, err := uc.Execute(context.Background(), CastVoteInput{
		MatchID: "test-match", VoterID: "p-1", VotedID: "p-2", Score: 4,
	})
	if !errors.Is(err, domainerrors.ErrMatchNotCompleted) {
		t.Errorf("expected ErrMatchNotCompleted, got %v", err)
	}
}

func TestCastVoteUseCase_Execute_ReturnsErrMatchNotFound_WhenMatchMissing(t *testing.T) {
	t.Parallel()
	uc := NewCastVoteUseCase(newFakeVoteRepository(), newFakeMatchRepository(),
		newFakeIDGenerator("v-1"), newFakeClock(time.Now()))
	_, err := uc.Execute(context.Background(), CastVoteInput{
		MatchID: "missing", VoterID: "p-1", VotedID: "p-2", Score: 4,
	})
	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestCastVoteUseCase_Execute_ReturnsErrSelfVote(t *testing.T) {
	t.Parallel()
	matchRepo := newFakeMatchRepository()
	seedCompletedMatch(t, matchRepo)
	uc := NewCastVoteUseCase(newFakeVoteRepository(), matchRepo,
		newFakeIDGenerator("v-1"), newFakeClock(time.Now()))

	_, err := uc.Execute(context.Background(), CastVoteInput{
		MatchID: "test-match", VoterID: "p-1", VotedID: "p-1", Score: 4,
	})
	if !errors.Is(err, domainerrors.ErrSelfVote) {
		t.Errorf("expected ErrSelfVote, got %v", err)
	}
}

func TestCastVoteUseCase_Execute_ReturnsErrAlreadyVoted_WhenDuplicate(t *testing.T) {
	t.Parallel()
	voteRepo := newFakeVoteRepository()
	matchRepo := newFakeMatchRepository()
	seedCompletedMatch(t, matchRepo)
	uc := NewCastVoteUseCase(voteRepo, matchRepo,
		newFakeIDGenerator("v-1", "v-2"), newFakeClock(time.Now()))

	_, err := uc.Execute(context.Background(), CastVoteInput{
		MatchID: "test-match", VoterID: "p-1", VotedID: "p-2", Score: 4,
	})
	if err != nil {
		t.Fatalf("first vote unexpected error: %v", err)
	}

	_, err = uc.Execute(context.Background(), CastVoteInput{
		MatchID: "test-match", VoterID: "p-1", VotedID: "p-2", Score: 5,
	})
	if !errors.Is(err, domainerrors.ErrAlreadyVoted) {
		t.Errorf("expected ErrAlreadyVoted, got %v", err)
	}
}
