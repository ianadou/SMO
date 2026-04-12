package invitation

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type fakeInvitationRepository struct {
	mu          sync.Mutex
	invitations map[entities.InvitationID]*entities.Invitation
}

func newFakeInvitationRepository() *fakeInvitationRepository {
	return &fakeInvitationRepository{invitations: make(map[entities.InvitationID]*entities.Invitation)}
}

func (r *fakeInvitationRepository) Save(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvitationRepository) FindByID(_ context.Context, id entities.InvitationID) (*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.invitations[id]
	if !ok {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *fakeInvitationRepository) FindByTokenHash(_ context.Context, hash string) (*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, inv := range r.invitations {
		if inv.TokenHash() == hash {
			return inv, nil
		}
	}
	return nil, domainerrors.ErrInvitationNotFound
}

func (r *fakeInvitationRepository) ListByMatch(_ context.Context, matchID entities.MatchID) ([]*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*entities.Invitation, 0)
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *fakeInvitationRepository) MarkAsUsed(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.invitations[inv.ID()]; !ok {
		return domainerrors.ErrInvitationNotFound
	}
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvitationRepository) Delete(_ context.Context, id entities.InvitationID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.invitations, id)
	return nil
}
