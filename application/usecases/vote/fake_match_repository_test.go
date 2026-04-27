package vote

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeMatchRepository is a minimal MatchRepository for Vote tests.
// Only FindByID is needed; other methods panic if accidentally called.
type fakeMatchRepository struct {
	mu      sync.Mutex
	matches map[entities.MatchID]*entities.Match
}

func newFakeMatchRepository() *fakeMatchRepository {
	return &fakeMatchRepository{matches: make(map[entities.MatchID]*entities.Match)}
}

func (r *fakeMatchRepository) addMatch(m *entities.Match) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[m.ID()] = m
}

func (r *fakeMatchRepository) FindByID(_ context.Context, id entities.MatchID) (*entities.Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.matches[id]
	if !ok {
		return nil, domainerrors.ErrMatchNotFound
	}
	return m, nil
}

func (r *fakeMatchRepository) Save(context.Context, *entities.Match) error {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) ListByGroup(context.Context, entities.GroupID) ([]*entities.Match, error) {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) UpdateStatus(context.Context, *entities.Match) error {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) Finalize(context.Context, *entities.Match) error {
	panic("not implemented in vote tests")
}

func (r *fakeMatchRepository) Delete(context.Context, entities.MatchID) error {
	panic("not implemented in vote tests")
}
