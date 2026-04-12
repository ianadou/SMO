package player

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// ListPlayersByGroupUseCase lists all players of a given group,
// ordered by ranking descending (handled by the repository).
type ListPlayersByGroupUseCase struct {
	playerRepo ports.PlayerRepository
}

// NewListPlayersByGroupUseCase builds a ListPlayersByGroupUseCase.
func NewListPlayersByGroupUseCase(playerRepo ports.PlayerRepository) *ListPlayersByGroupUseCase {
	return &ListPlayersByGroupUseCase{playerRepo: playerRepo}
}

// Execute returns all players that belong to the given group.
func (uc *ListPlayersByGroupUseCase) Execute(ctx context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	players, err := uc.playerRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list players by group use case: list for group %q: %w", groupID, err)
	}
	return players, nil
}
