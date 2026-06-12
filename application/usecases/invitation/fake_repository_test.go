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

	// persistedResponse records the last response actually committed
	// per invitation. The use case mutates the *entities.Invitation in
	// place before calling RespondWithCapacityGuard, so the entity
	// alone cannot tell us the pre-call stored state — this map is the
	// fake's equivalent of the row the postgres adapter re-reads inside
	// its transaction. It also drives the confirmed count so the
	// capacity check sees committed state, exactly like SQL COUNT does.
	persistedResponse map[entities.InvitationID]entities.InvitationResponse

	// Optional per-method error injectors for use case error-branch
	// tests. nil = method behaves normally; non-nil = method returns
	// the configured error verbatim.
	saveErr            error
	findByIDErr        error
	findByTokenHashErr error
	listByMatchErr     error
	countErr           error
	listConfirmedErr   error
	respondErr         error
}

func newFakeInvitationRepository() *fakeInvitationRepository {
	return &fakeInvitationRepository{
		invitations:       make(map[entities.InvitationID]*entities.Invitation),
		persistedResponse: make(map[entities.InvitationID]entities.InvitationResponse),
	}
}

func (r *fakeInvitationRepository) Save(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.invitations[inv.ID()] = inv
	r.persistedResponse[inv.ID()] = inv.Response()
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
	return r.countConfirmedLocked(matchID), nil
}

// countConfirmedLocked counts invitations whose committed response is
// 'yes' for a match (committed, not in-flight — like SQL COUNT). The
// caller must already hold r.mu.
func (r *fakeInvitationRepository) countConfirmedLocked(matchID entities.MatchID) int {
	count := 0
	for id, inv := range r.invitations {
		if inv.MatchID() == matchID && r.persistedResponse[id] == entities.InvitationResponseYes {
			count++
		}
	}
	return count
}

func (r *fakeInvitationRepository) ListConfirmedParticipants(_ context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listConfirmedErr != nil {
		return nil, r.listConfirmedErr
	}
	out := make([]entities.MatchParticipant, 0)
	for id, inv := range r.invitations {
		if inv.MatchID() == matchID && r.persistedResponse[id] == entities.InvitationResponseYes {
			out = append(out, entities.MatchParticipant{
				PlayerID:    inv.PlayerID(),
				PlayerName:  "Fake " + string(inv.PlayerID()),
				ConfirmedAt: *inv.RespondedAt(),
			})
		}
	}
	return out, nil
}

// Claim upserts without the adapter's conditional stored-row check:
// the fake hands out shared pointers, so the stored copy is already
// the claimed entity by the time Claim runs. No invitation use case
// in this package calls it; it exists to satisfy the port.
func (r *fakeInvitationRepository) Claim(_ context.Context, inv *entities.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invitations[inv.ID()] = inv
	return nil
}

// RespondWithCapacityGuard mirrors the postgres adapter's contract: the
// capacity check fires only when this call newly confirms an invitation
// (it was not already 'yes' in the stored map), and the whole thing is
// serialized under the fake's mutex to emulate the FOR UPDATE lock.
func (r *fakeInvitationRepository) RespondWithCapacityGuard(_ context.Context, inv *entities.Invitation, maxConfirmed int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.respondErr != nil {
		return r.respondErr
	}

	if _, ok := r.invitations[inv.ID()]; !ok {
		return domainerrors.ErrInvitationNotFound
	}
	previouslyConfirmed := r.persistedResponse[inv.ID()] == entities.InvitationResponseYes

	if inv.IsConfirmed() && !previouslyConfirmed {
		if r.countConfirmedLocked(inv.MatchID()) >= maxConfirmed {
			return domainerrors.ErrMatchFull
		}
	}

	r.invitations[inv.ID()] = inv
	r.persistedResponse[inv.ID()] = inv.Response()
	return nil
}

func (r *fakeInvitationRepository) Delete(_ context.Context, id entities.InvitationID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.invitations, id)
	return nil
}
