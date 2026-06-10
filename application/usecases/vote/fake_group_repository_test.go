package vote

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeGroupRepository is a minimal GroupRepository for vote tests: only
// FindByID is implemented, everything else panics if accidentally
// called.
type fakeGroupRepository struct {
	mu     sync.Mutex
	groups map[entities.GroupID]*entities.Group
}

func newFakeGroupRepository() *fakeGroupRepository {
	return &fakeGroupRepository{groups: make(map[entities.GroupID]*entities.Group)}
}

func (r *fakeGroupRepository) addGroup(g *entities.Group) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[g.ID()] = g
}

func (r *fakeGroupRepository) FindByID(_ context.Context, id entities.GroupID) (*entities.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.groups[id]
	if !ok {
		return nil, domainerrors.ErrGroupNotFound
	}
	return g, nil
}

func (r *fakeGroupRepository) Save(context.Context, *entities.Group) error {
	panic("not implemented in vote tests")
}

func (r *fakeGroupRepository) Update(context.Context, *entities.Group) error {
	panic("not implemented in vote tests")
}

func (r *fakeGroupRepository) ListByOrganizer(context.Context, entities.OrganizerID) ([]*entities.Group, error) {
	panic("not implemented in vote tests")
}

func (r *fakeGroupRepository) Delete(context.Context, entities.GroupID) error {
	panic("not implemented in vote tests")
}
