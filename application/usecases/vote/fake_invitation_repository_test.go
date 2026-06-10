package vote

import (
	"context"
	"sync"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeInvitationRepository is a minimal InvitationRepository for vote
// tests: only the token-hash lookup is implemented, everything else
// panics if accidentally called.
type fakeInvitationRepository struct {
	mu          sync.Mutex
	invitations map[string]*entities.Invitation
}

func newFakeInvitationRepository() *fakeInvitationRepository {
	return &fakeInvitationRepository{invitations: make(map[string]*entities.Invitation)}
}

func (r *fakeInvitationRepository) addInvitation(inv *entities.Invitation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invitations[inv.TokenHash()] = inv
}

func (r *fakeInvitationRepository) FindByTokenHash(_ context.Context, tokenHash string) (*entities.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.invitations[tokenHash]
	if !ok {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *fakeInvitationRepository) Save(context.Context, *entities.Invitation) error {
	panic("not implemented in vote tests")
}

func (r *fakeInvitationRepository) FindByID(context.Context, entities.InvitationID) (*entities.Invitation, error) {
	panic("not implemented in vote tests")
}

func (r *fakeInvitationRepository) ListByMatch(context.Context, entities.MatchID) ([]*entities.Invitation, error) {
	panic("not implemented in vote tests")
}

func (r *fakeInvitationRepository) CountConfirmedByMatch(context.Context, entities.MatchID) (int, error) {
	panic("not implemented in vote tests")
}

func (r *fakeInvitationRepository) ListConfirmedParticipants(context.Context, entities.MatchID) ([]entities.MatchParticipant, error) {
	panic("not implemented in vote tests")
}

func (r *fakeInvitationRepository) RespondWithCapacityGuard(context.Context, *entities.Invitation, int) error {
	panic("not implemented in vote tests")
}

func (r *fakeInvitationRepository) Delete(context.Context, entities.InvitationID) error {
	panic("not implemented in vote tests")
}
