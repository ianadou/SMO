package ports

import (
	"context"

	"github.com/ianadou/smo/domain/entities"
)

// MatchRepository is the persistence contract for the Match aggregate.
//
// Implementations live in the infrastructure layer (e.g., a PostgreSQL
// adapter using sqlc). The domain only depends on this interface, never
// on concrete database code.
type MatchRepository interface {
	// Save persists a new match.
	Save(ctx context.Context, match *entities.Match) error

	// FindByID looks up a match by its identifier. Returns
	// errors.ErrMatchNotFound (wrapped) if no match exists with that id.
	FindByID(ctx context.Context, id entities.MatchID) (*entities.Match, error)

	// ListByGroup returns all matches belonging to the given group,
	// ordered by scheduled date descending.
	ListByGroup(ctx context.Context, groupID entities.GroupID) ([]*entities.Match, error)

	// UpdateStatus persists a new status for the given match. The state
	// machine on the Match entity is responsible for validating which
	// transitions are allowed; this port trusts the caller.
	UpdateStatus(ctx context.Context, match *entities.Match) error

	// Finalize persists both the MVP and the new status (typically
	// closed) of the given match in a single update. Used by the
	// FinalizeMatchUseCase to avoid a partial state where MVP is
	// recorded but status hasn't transitioned yet.
	Finalize(ctx context.Context, match *entities.Match) error

	// Delete removes a match by its identifier.
	Delete(ctx context.Context, id entities.MatchID) error
}
