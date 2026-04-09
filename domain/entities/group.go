package entities

import (
	"strings"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const maxGroupNameLength = 100

// GroupID is the unique identifier of a Group.
type GroupID string

// OrganizerID is the unique identifier of an Organizer.
// Defined here because Group references it; the Organizer entity is
// declared in its own file.
type OrganizerID string

// Group represents a collection of players that play matches together.
// A group is owned by exactly one Organizer.
type Group struct {
	id          GroupID
	name        string
	organizerID OrganizerID
	createdAt   time.Time
}

// NewGroup builds a Group after validating its inputs.
func NewGroup(id GroupID, name string, organizerID OrganizerID, createdAt time.Time) (*Group, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > maxGroupNameLength {
		return nil, domainerrors.ErrInvalidName
	}

	if organizerID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	return &Group{
		id:          id,
		name:        trimmedName,
		organizerID: organizerID,
		createdAt:   createdAt,
	}, nil
}

// ID returns the group identifier.
func (g *Group) ID() GroupID { return g.id }

// Name returns the group name.
func (g *Group) Name() string { return g.name }

// OrganizerID returns the identifier of the organizer who owns this group.
func (g *Group) OrganizerID() OrganizerID { return g.organizerID }

// CreatedAt returns the creation timestamp of the group.
func (g *Group) CreatedAt() time.Time { return g.createdAt }
