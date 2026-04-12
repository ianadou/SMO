package ports

import (
	"context"

	"github.com/ianadou/smo/domain/entities"
)

// PlayerRepository is the persistence contract for the Player aggregate.
type PlayerRepository interface {
	Save(ctx context.Context, player *entities.Player) error
	FindByID(ctx context.Context, id entities.PlayerID) (*entities.Player, error)
	ListByGroup(ctx context.Context, groupID entities.GroupID) ([]*entities.Player, error)
	UpdateRanking(ctx context.Context, player *entities.Player) error
	Delete(ctx context.Context, id entities.PlayerID) error
}
