package ports

import (
	"context"

	"github.com/ianadou/smo/domain/entities"
)

// InvitationRepository is the persistence contract for the Invitation aggregate.
type InvitationRepository interface {
	Save(ctx context.Context, inv *entities.Invitation) error
	FindByID(ctx context.Context, id entities.InvitationID) (*entities.Invitation, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*entities.Invitation, error)
	ListByMatch(ctx context.Context, matchID entities.MatchID) ([]*entities.Invitation, error)

	// CountConfirmedByMatch returns the number of invitations for the
	// given match whose response is "yes" (i.e., participants who have
	// confirmed their attendance). Used to enforce
	// MaxParticipantsPerMatch.
	CountConfirmedByMatch(ctx context.Context, matchID entities.MatchID) (int, error)

	// ListConfirmedParticipants returns the flattened (player +
	// respondedAt) projection of every confirmed invitation
	// (response = "yes") for the given match, ordered by confirmation
	// time ascending. Used by the ListMatchParticipantsUseCase that
	// powers the match detail page.
	ListConfirmedParticipants(ctx context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error)

	// RespondWithCapacityGuard persists the invitation's new response.
	// When (and only when) the response transitions into "yes", it
	// atomically enforces the capacity cap: it serializes on the match
	// row, recounts confirmed invitations, and returns
	// domainerrors.ErrMatchFull if accepting this one would exceed
	// maxConfirmed. Changing to "no" (or re-confirming an invitation
	// that was already "yes") never triggers the capacity check.
	RespondWithCapacityGuard(ctx context.Context, inv *entities.Invitation, maxConfirmed int) error

	Delete(ctx context.Context, id entities.InvitationID) error
}
