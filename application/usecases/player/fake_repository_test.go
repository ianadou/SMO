package player

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type fakePlayerRepository struct {
	mu      sync.Mutex
	players map[entities.PlayerID]*entities.Player
}

func newFakePlayerRepository() *fakePlayerRepository {
	return &fakePlayerRepository{players: make(map[entities.PlayerID]*entities.Player)}
}

func (r *fakePlayerRepository) Save(_ context.Context, p *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players[p.ID()] = p
	return nil
}

func (r *fakePlayerRepository) FindByID(_ context.Context, id entities.PlayerID) (*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.players[id]
	if !ok {
		return nil, domainerrors.ErrPlayerNotFound
	}
	return p, nil
}

func (r *fakePlayerRepository) ListByGroup(_ context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Player, 0)
	for _, p := range r.players {
		if p.GroupID() == groupID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *fakePlayerRepository) UpdateRanking(_ context.Context, p *entities.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.players[p.ID()]; !ok {
		return domainerrors.ErrPlayerNotFound
	}
	r.players[p.ID()] = p
	return nil
}

func (r *fakePlayerRepository) Delete(_ context.Context, id entities.PlayerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.players, id)
	return nil
}
