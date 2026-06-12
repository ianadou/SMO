package ports

import (
	"context"

	"github.com/ianadou/smo/domain/entities"
)

// MatchShareLinkRepository is the persistence contract for the MatchShareLink aggregate.
type MatchShareLinkRepository interface {
	Create(ctx context.Context, link *entities.MatchShareLink) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entities.MatchShareLink, error)

	// FindActiveByMatchID returns the single link of the match that is
	// neither revoked nor expired, or ErrShareLinkNotFound when none
	// exists. Generating a link revokes the previous one, so at most
	// one active link per match is an invariant the adapter can rely on.
	FindActiveByMatchID(ctx context.Context, matchID entities.MatchID) (*entities.MatchShareLink, error)

	// Update persists the mutable state of the link (its revocation).
	Update(ctx context.Context, link *entities.MatchShareLink) error
}
