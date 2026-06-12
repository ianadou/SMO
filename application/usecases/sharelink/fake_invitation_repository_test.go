package sharelink

import (
	"context"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// fakeInvitationRepository stores invitations in a map. Save upserts
// (minting via join), and Claim upserts too: the conditional
// stored-row check of the real adapter is pointless here because the
// fake hands out shared pointers, so by claim time the stored copy is
// already the mutated entity. The race semantics belong to the
// postgres adapter's integration tests.
type fakeInvitationRepository struct {
	invitations map[entities.InvitationID]*entities.Invitation
	players     *fakePlayerRepository

	// Optional per-method error injectors for use case error-branch
	// tests. nil = method behaves normally.
	saveErr          error
	claimErr         error
	listByMatchErr   error
	listConfirmedErr error
}

func newFakeInvitationRepository(players *fakePlayerRepository) *fakeInvitationRepository {
	return &fakeInvitationRepository{
		invitations: make(map[entities.InvitationID]*entities.Invitation),
		players:     players,
	}
}

// seedInvitation seeds an invitation on match "match-1". The match id
// is fixed because every test in the package revolves around that
// single match; only the invitation's own state varies.
func (r *fakeInvitationRepository) seedInvitation(
	t testHelper,
	id entities.InvitationID,
	playerID entities.PlayerID,
	tokenHash string,
	expiresAt time.Time,
	response entities.InvitationResponse,
	respondedAt *time.Time,
	claimedAt *time.Time,
	createdAt time.Time,
) *entities.Invitation {
	inv, err := entities.NewInvitation(
		id, "match-1", playerID, tokenHash, expiresAt,
		response, respondedAt, claimedAt, createdAt,
	)
	if err != nil {
		t.Fatalf("seedInvitation: %v", err)
	}
	r.invitations[id] = inv
	return inv
}

func (r *fakeInvitationRepository) Save(_ context.Context, inv *entities.Invitation) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvitationRepository) Claim(_ context.Context, inv *entities.Invitation) error {
	if r.claimErr != nil {
		return r.claimErr
	}
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvitationRepository) FindByID(_ context.Context, id entities.InvitationID) (*entities.Invitation, error) {
	inv, ok := r.invitations[id]
	if !ok {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return inv, nil
}

func (r *fakeInvitationRepository) FindByTokenHash(_ context.Context, hash string) (*entities.Invitation, error) {
	for _, inv := range r.invitations {
		if inv.TokenHash() == hash {
			return inv, nil
		}
	}
	return nil, domainerrors.ErrInvitationNotFound
}

func (r *fakeInvitationRepository) ListByMatch(_ context.Context, matchID entities.MatchID) ([]*entities.Invitation, error) {
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
	count := 0
	for _, inv := range r.invitations {
		if inv.MatchID() == matchID && inv.IsConfirmed() {
			count++
		}
	}
	return count, nil
}

func (r *fakeInvitationRepository) ListConfirmedParticipants(_ context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error) {
	if r.listConfirmedErr != nil {
		return nil, r.listConfirmedErr
	}
	out := make([]entities.MatchParticipant, 0)
	for _, inv := range r.invitations {
		if inv.MatchID() != matchID || !inv.IsConfirmed() {
			continue
		}
		name := string(inv.PlayerID())
		if p, ok := r.players.players[inv.PlayerID()]; ok {
			name = p.Name()
		}
		out = append(out, entities.MatchParticipant{
			PlayerID:    inv.PlayerID(),
			PlayerName:  name,
			ConfirmedAt: *inv.RespondedAt(),
		})
	}
	return out, nil
}

func (r *fakeInvitationRepository) RespondWithCapacityGuard(_ context.Context, inv *entities.Invitation, _ int) error {
	r.invitations[inv.ID()] = inv
	return nil
}

func (r *fakeInvitationRepository) Delete(_ context.Context, id entities.InvitationID) error {
	delete(r.invitations, id)
	return nil
}
