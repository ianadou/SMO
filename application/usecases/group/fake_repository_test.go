package group

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeGroupRepository is an in-memory implementation of the
// GroupRepository port used by use case unit tests.
//
// It is goroutine-safe so tests can run in parallel without races.
type fakeGroupRepository struct {
	mu     sync.Mutex
	groups map[entities.GroupID]*entities.Group
}

func newFakeGroupRepository() *fakeGroupRepository {
	return &fakeGroupRepository{
		groups: make(map[entities.GroupID]*entities.Group),
	}
}

func (r *fakeGroupRepository) Save(_ context.Context, group *entities.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[group.ID()] = group
	return nil
}

func (r *fakeGroupRepository) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	group, exists := r.groups[id]
	if !exists {
		return nil, domainerrors.ErrGroupNotFound
	}
	return group, nil
}

func (r *fakeGroupRepository) ListByOrganizer(_ context.Context, organizerID entities.OrganizerID) ([]*entities.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Group, 0)
	for _, group := range r.groups {
		if group.OrganizerID() == organizerID {
			result = append(result, group)
		}
	}
	return result, nil
}

func (r *fakeGroupRepository) Update(_ context.Context, group *entities.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID()]; !exists {
		return domainerrors.ErrGroupNotFound
	}
	r.groups[group.ID()] = group
	return nil
}

func (r *fakeGroupRepository) Delete(_ context.Context, id entities.GroupID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.groups, id)
	return nil
}
