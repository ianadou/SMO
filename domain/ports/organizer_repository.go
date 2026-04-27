package ports

import (
	"context"

	"github.com/ianadou/smo/domain/entities"
)

// OrganizerRepository is the persistence contract for the Organizer
// aggregate.
type OrganizerRepository interface {
	// Save persists a new organizer. Returns ErrEmailAlreadyExists
	// (wrapped) if the unique constraint on email is violated.
	Save(ctx context.Context, organizer *entities.Organizer) error

	// FindByID looks up an organizer by ID. Returns ErrOrganizerNotFound
	// (wrapped) if no match.
	FindByID(ctx context.Context, id entities.OrganizerID) (*entities.Organizer, error)

	// FindByEmail looks up an organizer by email. Returns
	// ErrOrganizerNotFound (wrapped) if no match. Used by the login
	// flow; callers must translate this into ErrInvalidCredentials at
	// the use case level to avoid email enumeration.
	FindByEmail(ctx context.Context, email string) (*entities.Organizer, error)
}
