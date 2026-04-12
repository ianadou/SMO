package invitation

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// ListInvitationsByMatchUseCase lists invitations belonging to a match.
type ListInvitationsByMatchUseCase struct {
	repo ports.InvitationRepository
}

// NewListInvitationsByMatchUseCase builds the use case.
func NewListInvitationsByMatchUseCase(repo ports.InvitationRepository) *ListInvitationsByMatchUseCase {
	return &ListInvitationsByMatchUseCase{repo: repo}
}

// Execute returns all invitations that belong to the given match.
func (uc *ListInvitationsByMatchUseCase) Execute(ctx context.Context, matchID entities.MatchID) ([]*entities.Invitation, error) {
	invitations, err := uc.repo.ListByMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("list invitations by match use case: list for %q: %w", matchID, err)
	}
	return invitations, nil
}
