package inmemory

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// GroupRepository is an in-memory implementation of the
// domain ports.GroupRepository interface, suitable for local
// development and tests.
//
// It uses a mutex to be safe for concurrent access from multiple
// goroutines, which matches the production behavior of HTTP servers
// handling parallel requests.
type GroupRepository struct {
	mu     sync.RWMutex
	groups map[entities.GroupID]*entities.Group
}

// NewGroupRepository builds an empty in-memory GroupRepository.
func NewGroupRepository() *GroupRepository {
	return &GroupRepository{
		groups: make(map[entities.GroupID]*entities.Group),
	}
}

// Save stores a group, overwriting any previous group with the same ID.
func (r *GroupRepository) Save(_ context.Context, group *entities.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[group.ID()] = group
	return nil
}

// FindByID returns the group with the given ID or ErrGroupNotFound.
func (r *GroupRepository) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	group, exists := r.groups[id]
	if !exists {
		return nil, domainerrors.ErrGroupNotFound
	}
	return group, nil
}

// ListByOrganizer returns all groups owned by the given organizer.
func (r *GroupRepository) ListByOrganizer(_ context.Context, organizerID entities.OrganizerID) ([]*entities.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entities.Group, 0)
	for _, group := range r.groups {
		if group.OrganizerID() == organizerID {
			result = append(result, group)
		}
	}
	return result, nil
}

// Update overwrites an existing group. Returns ErrGroupNotFound if no
// group exists with that ID.
func (r *GroupRepository) Update(_ context.Context, group *entities.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID()]; !exists {
		return domainerrors.ErrGroupNotFound
	}
	r.groups[group.ID()] = group
	return nil
}

// Delete removes the group with the given ID. Idempotent: deleting a
// non-existent ID is a no-op.
func (r *GroupRepository) Delete(_ context.Context, id entities.GroupID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.groups, id)
	return nil
}
