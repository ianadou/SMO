package player

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// GetPlayerUseCase retrieves a Player by identifier.
type GetPlayerUseCase struct {
	playerRepo ports.PlayerRepository
}

// NewGetPlayerUseCase builds a GetPlayerUseCase.
func NewGetPlayerUseCase(playerRepo ports.PlayerRepository) *GetPlayerUseCase {
	return &GetPlayerUseCase{playerRepo: playerRepo}
}

// Execute returns the player with the given ID.
func (uc *GetPlayerUseCase) Execute(ctx context.Context, id entities.PlayerID) (*entities.Player, error) {
	player, err := uc.playerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get player use case: find player %q: %w", id, err)
	}
	return player, nil
}
