package invitation

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// AcceptInvitationUseCase consumes an invitation by its plain token.
// The workflow is:
//
//  1. hash the plain token
//  2. look up the invitation by hash
//  3. count confirmed invitations for the same match; reject with
//     ErrMatchFull if the count already reached MaxParticipantsPerMatch
//     (FCFS policy — see ADR 0008)
//  4. MarkAsUsed (validates not-used / not-expired on the entity)
//  5. persist
type AcceptInvitationUseCase struct {
	repo   ports.InvitationRepository
	tokens ports.InvitationTokenService
	clock  ports.Clock
}

// NewAcceptInvitationUseCase builds the use case.
func NewAcceptInvitationUseCase(
	repo ports.InvitationRepository,
	tokens ports.InvitationTokenService,
	clock ports.Clock,
) *AcceptInvitationUseCase {
	return &AcceptInvitationUseCase{repo: repo, tokens: tokens, clock: clock}
}

// Execute accepts the invitation identified by the given plain token.
// Returns ErrInvitationNotFound if no invitation matches the hash,
// ErrInvitationExpired / ErrInvitationAlreadyUsed if the invitation
// cannot be consumed, or ErrMatchFull if the match already reached
// MaxParticipantsPerMatch confirmed invitations.
func (uc *AcceptInvitationUseCase) Execute(ctx context.Context, plainToken string) (*entities.Invitation, error) {
	hash := uc.tokens.HashToken(plainToken)

	inv, err := uc.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("accept invitation use case: find by hash: %w", err)
	}

	confirmed, err := uc.repo.CountConfirmedByMatch(ctx, inv.MatchID())
	if err != nil {
		return nil, fmt.Errorf("accept invitation use case: count confirmed: %w", err)
	}
	if confirmed >= entities.MaxParticipantsPerMatch {
		return nil, fmt.Errorf("accept invitation use case: %w", domainerrors.ErrMatchFull)
	}

	if markErr := inv.MarkAsUsed(uc.clock.Now()); markErr != nil {
		return nil, fmt.Errorf("accept invitation use case: mark used: %w", markErr)
	}

	if saveErr := uc.repo.MarkAsUsed(ctx, inv); saveErr != nil {
		return nil, fmt.Errorf("accept invitation use case: persist: %w", saveErr)
	}

	return inv, nil
}
