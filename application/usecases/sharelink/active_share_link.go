package sharelink

import (
	"context"
	"fmt"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// findActiveShareLink resolves a bearer-submitted plain token into the
// active share link it designates. An unknown hash surfaces as
// ErrShareLinkNotFound; a link that exists but is revoked or expired
// surfaces as ErrShareLinkInactive. The two sentinels stay distinct for
// logs and metrics, but the HTTP layer maps both to the same 404 so a
// token probe cannot learn which way the link died.
//
// Errors carry step context but no use case prefix: each caller wraps
// the result with its own name.
func findActiveShareLink(
	ctx context.Context,
	links ports.MatchShareLinkRepository,
	tokens ports.InvitationTokenService,
	plainToken string,
	now time.Time,
) (*entities.MatchShareLink, error) {
	link, err := links.FindByTokenHash(ctx, tokens.HashToken(plainToken))
	if err != nil {
		return nil, fmt.Errorf("find link by hash: %w", err)
	}

	if !link.IsActive(now) {
		return nil, fmt.Errorf("link %q is dead: %w",
			link.ID(), domainerrors.ErrShareLinkInactive)
	}

	return link, nil
}
