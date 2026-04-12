package match

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// GetMatchUseCase retrieves a Match by its identifier.
type GetMatchUseCase struct {
	matchRepo ports.MatchRepository
}

// NewGetMatchUseCase builds a GetMatchUseCase.
func NewGetMatchUseCase(matchRepo ports.MatchRepository) *GetMatchUseCase {
	return &GetMatchUseCase{matchRepo: matchRepo}
}

// Execute returns the match with the given ID.
func (uc *GetMatchUseCase) Execute(ctx context.Context, id entities.MatchID) (*entities.Match, error) {
	match, err := uc.matchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get match use case: find match %q: %w", id, err)
	}
	return match, nil
}
