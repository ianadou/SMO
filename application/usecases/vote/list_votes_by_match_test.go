package vote

import (
	"context"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestListVotesByMatchUseCase_Execute_ReturnsAllVotesForMatch(t *testing.T) {
	t.Parallel()
	repo := newFakeVoteRepository()
	ctx := context.Background()

	// 2 votes on match-1 (different voted), 1 on match-2 to verify isolation.
	v1, _ := entities.NewVote("v-1", "match-1", "p-1", "p-2", 4, time.Now())
	v2, _ := entities.NewVote("v-2", "match-1", "p-1", "p-3", 5, time.Now())
	v3, _ := entities.NewVote("v-3", "match-2", "p-1", "p-2", 3, time.Now())
	_ = repo.Save(ctx, v1)
	_ = repo.Save(ctx, v2)
	_ = repo.Save(ctx, v3)

	uc := NewListVotesByMatchUseCase(repo)
	votes, err := uc.Execute(ctx, "match-1")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(votes) != 2 {
		t.Errorf("expected 2 votes for match-1, got %d", len(votes))
	}
}
