package sharelink

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// RevokeMatchShareLinkUseCase kills the active share link of a match on
// behalf of its organizer, so the URL circulating in the group chat
// stops resolving.
type RevokeMatchShareLinkUseCase struct {
	links     ports.MatchShareLinkRepository
	matchRepo ports.MatchRepository
	groupRepo ports.GroupRepository
	clock     ports.Clock
}

// NewRevokeMatchShareLinkUseCase builds the use case.
func NewRevokeMatchShareLinkUseCase(
	links ports.MatchShareLinkRepository,
	matchRepo ports.MatchRepository,
	groupRepo ports.GroupRepository,
	clock ports.Clock,
) *RevokeMatchShareLinkUseCase {
	return &RevokeMatchShareLinkUseCase{
		links:     links,
		matchRepo: matchRepo,
		groupRepo: groupRepo,
		clock:     clock,
	}
}

// Execute revokes the active link of the match. Returns ErrMatchNotFound
// for an unknown OR foreign match, and ErrShareLinkNotFound when the
// match has no active link to revoke.
func (uc *RevokeMatchShareLinkUseCase) Execute(
	ctx context.Context,
	matchID entities.MatchID,
	organizerID entities.OrganizerID,
) error {
	if err := verifyMatchOwnership(ctx, uc.matchRepo, uc.groupRepo, matchID, organizerID); err != nil {
		return fmt.Errorf("revoke match share link use case: %w", err)
	}

	link, err := uc.links.FindActiveByMatchID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("revoke match share link use case: find active link: %w", err)
	}

	if revokeErr := link.Revoke(uc.clock.Now()); revokeErr != nil {
		return fmt.Errorf("revoke match share link use case: revoke: %w", revokeErr)
	}

	if updateErr := uc.links.Update(ctx, link); updateErr != nil {
		return fmt.Errorf("revoke match share link use case: persist: %w", updateErr)
	}

	return nil
}
