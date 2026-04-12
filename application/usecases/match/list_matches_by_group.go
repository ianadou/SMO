package match

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// ListMatchesByGroupUseCase retrieves all matches for a given group.
type ListMatchesByGroupUseCase struct {
	matchRepo ports.MatchRepository
}

// NewListMatchesByGroupUseCase builds a ListMatchesByGroupUseCase.
func NewListMatchesByGroupUseCase(matchRepo ports.MatchRepository) *ListMatchesByGroupUseCase {
	return &ListMatchesByGroupUseCase{matchRepo: matchRepo}
}

// Execute returns all matches that belong to the given group.
func (uc *ListMatchesByGroupUseCase) Execute(ctx context.Context, groupID entities.GroupID) ([]*entities.Match, error) {
	matches, err := uc.matchRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list matches by group use case: list for group %q: %w", groupID, err)
	}
	return matches, nil
}
