package vote

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// GetVoteUseCase retrieves a vote by ID.
type GetVoteUseCase struct {
	repo ports.VoteRepository
}

// NewGetVoteUseCase builds the use case.
func NewGetVoteUseCase(repo ports.VoteRepository) *GetVoteUseCase {
	return &GetVoteUseCase{repo: repo}
}

// Execute returns the vote with the given ID.
func (uc *GetVoteUseCase) Execute(ctx context.Context, id entities.VoteID) (*entities.Vote, error) {
	v, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get vote use case: find %q: %w", id, err)
	}
	return v, nil
}
