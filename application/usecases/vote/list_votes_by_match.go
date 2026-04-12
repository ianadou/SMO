package vote

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// ListVotesByMatchUseCase lists all votes cast during a match.
type ListVotesByMatchUseCase struct {
	repo ports.VoteRepository
}

// NewListVotesByMatchUseCase builds the use case.
func NewListVotesByMatchUseCase(repo ports.VoteRepository) *ListVotesByMatchUseCase {
	return &ListVotesByMatchUseCase{repo: repo}
}

// Execute returns the list of votes for the given match.
func (uc *ListVotesByMatchUseCase) Execute(ctx context.Context, matchID entities.MatchID) ([]*entities.Vote, error) {
	votes, err := uc.repo.ListByMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("list votes by match use case: list for %q: %w", matchID, err)
	}
	return votes, nil
}
