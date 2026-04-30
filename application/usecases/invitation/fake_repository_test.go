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

	// Optional per-method error injectors for use case error-branch
	// tests. nil = method behaves normally; non-nil = method returns
	// the configured error verbatim.
	saveErr            error
	findByIDErr        error
	findByTokenHashErr error
	listByMatchErr     error
	countErr           error
	listConfirmedErr   error
	markAsUsedErr      error
}

func newFakeInvitationRepository() *fakeInvitationRepository {
	return &fakeInvitationRepository{invitations: make(map[entities.InvitationID]*entities.Invitation)}
}

func (r *fakeInvitationRepository) Save(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvitationRepository) FindByID(_ context.Context, id entities.InvitationID) (*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.findByIDErr != nil {
		return nil, r.findByIDErr
	}
	inv, ok := r.invitations[id]
	if !ok {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *fakeInvitationRepository) FindByTokenHash(_ context.Context, hash string) (*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.findByTokenHashErr != nil {
		return nil, r.findByTokenHashErr
	}
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
	if r.listByMatchErr != nil {
		return nil, r.listByMatchErr
	}
	result := make([]*entities.Invitation, 0)
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *fakeInvitationRepository) CountConfirmedByMatch(_ context.Context, matchID entities.MatchID) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.countErr != nil {
		return 0, r.countErr
	}
	count := 0
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID && inv.IsUsed() {
			count++
		}
	}
	return count, nil
}

func (r *fakeInvitationRepository) ListConfirmedParticipants(_ context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listConfirmedErr != nil {
		return nil, r.listConfirmedErr
	}
	out := make([]entities.MatchParticipant, 0)
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID && inv.IsUsed() {
			out = append(out, entities.MatchParticipant{
				PlayerID:    inv.PlayerID(),
				PlayerName:  "Fake " + string(inv.PlayerID()),
				ConfirmedAt: *inv.UsedAt(),
			})
		}
	}
	return out, nil
}

func (r *fakeInvitationRepository) MarkAsUsed(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.markAsUsedErr != nil {
		return r.markAsUsedErr
	}
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
