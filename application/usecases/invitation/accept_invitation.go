package invitation

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// AcceptInvitationUseCase consumes an invitation by its plain token.
// The workflow is: hash the plain token → look up invitation by hash →
// MarkAsUsed (validates not used, not expired) → persist.
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
// cannot be consumed.
func (uc *AcceptInvitationUseCase) Execute(ctx context.Context, plainToken string) (*entities.Invitation, error) {
	hash := uc.tokens.HashToken(plainToken)

	inv, err := uc.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("accept invitation use case: find by hash: %w", err)
	}

	if markErr := inv.MarkAsUsed(uc.clock.Now()); markErr != nil {
		return nil, fmt.Errorf("accept invitation use case: mark used: %w", markErr)
	}

	if saveErr := uc.repo.MarkAsUsed(ctx, inv); saveErr != nil {
		return nil, fmt.Errorf("accept invitation use case: persist: %w", saveErr)
	}

	return inv, nil
}
