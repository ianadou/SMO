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
	// given match whose used_at is not NULL (i.e., participants who have
	// confirmed their attendance). Used by AcceptInvitationUseCase to
	// enforce MaxParticipantsPerMatch.
	CountConfirmedByMatch(ctx context.Context, matchID entities.MatchID) (int, error)

	// ListConfirmedParticipants returns the flattened (player + used_at)
	// projection of every confirmed invitation for the given match,
	// ordered by confirmation time ascending. Used by the
	// ListMatchParticipantsUseCase that powers the match detail page.
	ListConfirmedParticipants(ctx context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error)

	MarkAsUsed(ctx context.Context, inv *entities.Invitation) error
	Delete(ctx context.Context, id entities.InvitationID) error
}
