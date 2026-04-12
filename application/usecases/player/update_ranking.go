package player

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// UpdatePlayerRankingUseCase changes a player's ranking. This is the
// hook used by match-result processing to adjust rankings after votes
// have been tallied.
type UpdatePlayerRankingUseCase struct {
	playerRepo ports.PlayerRepository
}

// NewUpdatePlayerRankingUseCase builds the use case.
func NewUpdatePlayerRankingUseCase(playerRepo ports.PlayerRepository) *UpdatePlayerRankingUseCase {
	return &UpdatePlayerRankingUseCase{playerRepo: playerRepo}
}

// Execute replaces the player's ranking with the given new value. The
// player is loaded, its ranking is updated through NewPlayer (which
// validates invariants), and the new entity is persisted.
func (uc *UpdatePlayerRankingUseCase) Execute(ctx context.Context, id entities.PlayerID, newRanking int) (*entities.Player, error) {
	existing, err := uc.playerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("update player ranking use case: find player %q: %w", id, err)
	}

	// Rebuild the player with the new ranking. This goes through NewPlayer
	// so any invariant on name length etc. is re-validated, which matters
	// if the entity validation rules evolve.
	updated, buildErr := entities.NewPlayer(existing.ID(), existing.GroupID(), existing.Name(), newRanking)
	if buildErr != nil {
		return nil, fmt.Errorf("update player ranking use case: rebuild player %q: %w", id, buildErr)
	}

	if saveErr := uc.playerRepo.UpdateRanking(ctx, updated); saveErr != nil {
		return nil, fmt.Errorf("update player ranking use case: persist %q: %w", id, saveErr)
	}

	return updated, nil
}
