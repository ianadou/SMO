package ports

import (
	"context"

	"github.com/ianadou/smo/domain/entities"
)

// GroupRepository is the persistence contract for the Group aggregate.
//
// Implementations live in the infrastructure layer (e.g., a PostgreSQL
// adapter using sqlc). The domain only depends on this interface, never
// on concrete database code.
//
// All methods take a context.Context as their first parameter to support
// cancellation, deadlines, and request-scoped values such as tracing.
type GroupRepository interface {
	// Save persists a new group. Returns an error if the group cannot be
	// stored (e.g., the database is unreachable or a unique constraint
	// is violated).
	Save(ctx context.Context, group *entities.Group) error

	// FindByID looks up a group by its identifier. Returns
	// errors.ErrGroupNotFound (wrapped) if no group exists with that id.
	FindByID(ctx context.Context, id entities.GroupID) (*entities.Group, error)

	// ListByOrganizer returns all groups owned by the given organizer,
	// ordered by creation time descending (newest first).
	ListByOrganizer(ctx context.Context, organizerID entities.OrganizerID) ([]*entities.Group, error)

	// Update modifies an existing group. The group's ID is used to
	// locate the row; the other fields are overwritten with the values
	// from the entity. Returns errors.ErrGroupNotFound if no group exists
	// with that id.
	Update(ctx context.Context, group *entities.Group) error

	// Delete removes a group by its identifier. Returns
	// errors.ErrGroupNotFound if no group exists with that id.
	Delete(ctx context.Context, id entities.GroupID) error
}
