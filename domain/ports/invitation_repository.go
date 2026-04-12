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
	MarkAsUsed(ctx context.Context, inv *entities.Invitation) error
	Delete(ctx context.Context, id entities.InvitationID) error
}
