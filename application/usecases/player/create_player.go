package player

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// CreatePlayerUseCase orchestrates the creation of a new Player with
// the default starting ranking.
type CreatePlayerUseCase struct {
	playerRepo ports.PlayerRepository
	idGen      ports.IDGenerator
}

// CreatePlayerInput is the parameter struct for CreatePlayerUseCase.Execute.
type CreatePlayerInput struct {
	GroupID entities.GroupID
	Name    string
}

// NewCreatePlayerUseCase builds a CreatePlayerUseCase.
func NewCreatePlayerUseCase(playerRepo ports.PlayerRepository, idGen ports.IDGenerator) *CreatePlayerUseCase {
	return &CreatePlayerUseCase{playerRepo: playerRepo, idGen: idGen}
}

// Execute creates a new Player with the default ranking.
//
// The ranking is intentionally not part of the input: organizers should
// not pick arbitrary rankings at creation time. An admin path for
// importing rankings from an external source may be added later.
func (uc *CreatePlayerUseCase) Execute(ctx context.Context, input CreatePlayerInput) (*entities.Player, error) {
	id := entities.PlayerID(uc.idGen.Generate())

	player, err := entities.NewPlayer(id, input.GroupID, input.Name, entities.DefaultPlayerRanking())
	if err != nil {
		return nil, fmt.Errorf("create player use case: build player: %w", err)
	}

	if saveErr := uc.playerRepo.Save(ctx, player); saveErr != nil {
		return nil, fmt.Errorf("create player use case: save player %q: %w", player.ID(), saveErr)
	}

	return player, nil
}
