package invitation

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// RespondToInvitationUseCase records (or changes) a player's answer to
// an invitation identified by its plain token. The workflow is:
//
//  1. hash the plain token, look up the invitation by hash
//  2. load the match to decide whether attendance is locked
//     (locked from teams_ready onward)
//  3. apply the answer on the entity (enforces valid answer / not
//     expired / not locked)
//  4. persist through the capacity-guarded repository call, which
//     atomically rejects with ErrMatchFull if confirming would exceed
//     MaxParticipantsPerMatch (FCFS policy — see ADR 0008)
//
// Capacity is enforced only inside the repository transaction; the use
// case never pre-counts, which is what removes the original
// check-then-act race.
type RespondToInvitationUseCase struct {
	repo      ports.InvitationRepository
	matchRepo ports.MatchRepository
	tokens    ports.InvitationTokenService
	clock     ports.Clock
}

// NewRespondToInvitationUseCase builds the use case.
func NewRespondToInvitationUseCase(
	repo ports.InvitationRepository,
	matchRepo ports.MatchRepository,
	tokens ports.InvitationTokenService,
	clock ports.Clock,
) *RespondToInvitationUseCase {
	return &RespondToInvitationUseCase{
		repo:      repo,
		matchRepo: matchRepo,
		tokens:    tokens,
		clock:     clock,
	}
}

// Execute applies the answer to the invitation identified by the given
// plain token. Returns ErrInvitationNotFound if no invitation matches
// the hash, ErrInvalidInvitationResponse for an unsettable answer,
// ErrInvitationExpired / ErrInvitationLocked if the answer cannot be
// changed, or ErrMatchFull if confirming would exceed the participant
// cap.
func (uc *RespondToInvitationUseCase) Execute(
	ctx context.Context,
	plainToken string,
	answer entities.InvitationResponse,
) (*entities.Invitation, error) {
	hash := uc.tokens.HashToken(plainToken)

	inv, err := uc.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("respond to invitation use case: find by hash: %w", err)
	}

	match, err := uc.matchRepo.FindByID(ctx, inv.MatchID())
	if err != nil {
		return nil, fmt.Errorf("respond to invitation use case: find match: %w", err)
	}

	if respondErr := inv.Respond(answer, uc.clock.Now(), match.AttendanceLocked()); respondErr != nil {
		return nil, fmt.Errorf("respond to invitation use case: respond: %w", respondErr)
	}

	if saveErr := uc.repo.RespondWithCapacityGuard(ctx, inv, entities.MaxParticipantsPerMatch); saveErr != nil {
		return nil, fmt.Errorf("respond to invitation use case: persist: %w", saveErr)
	}

	return inv, nil
}
