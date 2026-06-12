package sharelink

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// verifyMatchOwnership checks that the requester owns the group the
// match belongs to. A match owned by another organizer is reported as
// ErrMatchNotFound rather than forbidden, so share-link endpoints never
// reveal whether a match ID exists (same anti-oracle rule as
// RenameGroupUseCase).
//
// Errors carry step context but no use case prefix: each caller wraps
// the result with its own name.
func verifyMatchOwnership(
	ctx context.Context,
	matchRepo ports.MatchRepository,
	groupRepo ports.GroupRepository,
	matchID entities.MatchID,
	organizerID entities.OrganizerID,
) error {
	match, err := matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("find match: %w", err)
	}

	group, err := groupRepo.FindByID(ctx, match.GroupID())
	if err != nil {
		return fmt.Errorf("find group: %w", err)
	}

	if group.OrganizerID() != organizerID {
		return fmt.Errorf("match %q not owned by requester: %w",
			matchID, domainerrors.ErrMatchNotFound)
	}

	return nil
}
