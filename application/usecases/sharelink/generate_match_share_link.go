package sharelink

import (
	"context"
	"errors"
	"fmt"

	"github.com/ianadou/smo/application/usecases/invitation"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// GenerateMatchShareLinkUseCase creates the shareable link of a match
// and returns both the stored entity (with the token hash) and the
// plain token. The plain token is available only in this call's return
// value; it is never stored or returned again.
//
// A match has at most one active link: any previously active link is
// revoked before the new one is created, so regenerating invalidates
// the URL already circulating in the group chat.
type GenerateMatchShareLinkUseCase struct {
	links     ports.MatchShareLinkRepository
	matchRepo ports.MatchRepository
	groupRepo ports.GroupRepository
	tokens    ports.InvitationTokenService
	idGen     ports.IDGenerator
	clock     ports.Clock
}

// GenerateMatchShareLinkResult bundles the saved link with the plain
// token. Callers must surface the token to the organizer exactly once.
type GenerateMatchShareLinkResult struct {
	ShareLink  *entities.MatchShareLink
	PlainToken string
}

// NewGenerateMatchShareLinkUseCase builds the use case.
func NewGenerateMatchShareLinkUseCase(
	links ports.MatchShareLinkRepository,
	matchRepo ports.MatchRepository,
	groupRepo ports.GroupRepository,
	tokens ports.InvitationTokenService,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *GenerateMatchShareLinkUseCase {
	return &GenerateMatchShareLinkUseCase{
		links:     links,
		matchRepo: matchRepo,
		groupRepo: groupRepo,
		tokens:    tokens,
		idGen:     idGen,
		clock:     clock,
	}
}

// Execute revokes the previous active link (if any), mints a new token,
// persists the link, and returns the plain token once. Returns
// ErrMatchNotFound for an unknown OR foreign match.
//
// The link expires after the same validity window as invitations: both
// artifacts gate entry to the same match, so they share one policy.
func (uc *GenerateMatchShareLinkUseCase) Execute(
	ctx context.Context,
	matchID entities.MatchID,
	organizerID entities.OrganizerID,
) (*GenerateMatchShareLinkResult, error) {
	if err := verifyMatchOwnership(ctx, uc.matchRepo, uc.groupRepo, matchID, organizerID); err != nil {
		return nil, fmt.Errorf("generate match share link use case: %w", err)
	}

	now := uc.clock.Now()

	previous, err := uc.links.FindActiveByMatchID(ctx, matchID)
	switch {
	case err == nil:
		if revokeErr := previous.Revoke(now); revokeErr != nil {
			return nil, fmt.Errorf("generate match share link use case: revoke previous: %w", revokeErr)
		}
		if updateErr := uc.links.Update(ctx, previous); updateErr != nil {
			return nil, fmt.Errorf("generate match share link use case: persist revocation: %w", updateErr)
		}
	case !errors.Is(err, domainerrors.ErrShareLinkNotFound):
		return nil, fmt.Errorf("generate match share link use case: find active link: %w", err)
	}

	plainToken, err := uc.tokens.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generate match share link use case: generate token: %w", err)
	}
	hash := uc.tokens.HashToken(plainToken)

	link, err := entities.NewMatchShareLink(
		entities.MatchShareLinkID(uc.idGen.Generate()),
		matchID,
		hash,
		now.Add(invitation.DefaultInvitationValidityDuration),
		nil,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("generate match share link use case: build link: %w", err)
	}

	if createErr := uc.links.Create(ctx, link); createErr != nil {
		return nil, fmt.Errorf("generate match share link use case: save %q: %w", link.ID(), createErr)
	}

	return &GenerateMatchShareLinkResult{ShareLink: link, PlainToken: plainToken}, nil
}
