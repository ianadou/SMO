package invitation

import (
	"context"
	"fmt"
	"time"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// DefaultInvitationValidityDuration is how long an invitation stays
// valid when the caller does not provide an explicit expires_at.
const DefaultInvitationValidityDuration = 5 * 24 * time.Hour

// CreateInvitationUseCase creates a new invitation and returns both the
// stored entity (with the token hash) and the plain token that the
// organizer must share with the invitee. The plain token is available
// only in this call's return value; it is never stored or returned again.
type CreateInvitationUseCase struct {
	repo   ports.InvitationRepository
	tokens ports.InvitationTokenService
	idGen  ports.IDGenerator
	clock  ports.Clock
}

// CreateInvitationInput is the input of Execute. ExpiresAt is optional:
// if zero, the default validity window is applied.
type CreateInvitationInput struct {
	MatchID   entities.MatchID
	ExpiresAt time.Time
}

// CreateInvitationResult bundles the saved invitation with the plain
// token. Callers must surface the token to the end user exactly once.
type CreateInvitationResult struct {
	Invitation *entities.Invitation
	PlainToken string
}

// NewCreateInvitationUseCase builds the use case.
func NewCreateInvitationUseCase(
	repo ports.InvitationRepository,
	tokens ports.InvitationTokenService,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *CreateInvitationUseCase {
	return &CreateInvitationUseCase{repo: repo, tokens: tokens, idGen: idGen, clock: clock}
}

// Execute generates a token, hashes it, persists the invitation, and
// returns the plain token once.
func (uc *CreateInvitationUseCase) Execute(ctx context.Context, input CreateInvitationInput) (*CreateInvitationResult, error) {
	now := uc.clock.Now()

	expiresAt := input.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = now.Add(DefaultInvitationValidityDuration)
	}

	plainToken, err := uc.tokens.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("create invitation use case: generate token: %w", err)
	}
	hash := uc.tokens.HashToken(plainToken)

	// PlayerID is wired through the input + use case in the next commit
	// (PR A C4); for now we pass a placeholder so the entity build does not
	// reject empty playerID. The placeholder is overwritten by C4 before
	// the change reaches main via squash-merge.
	inv, err := entities.NewInvitation(
		entities.InvitationID(uc.idGen.Generate()),
		input.MatchID,
		entities.PlayerID("p-placeholder"),
		hash,
		expiresAt,
		nil,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("create invitation use case: build invitation: %w", err)
	}

	if saveErr := uc.repo.Save(ctx, inv); saveErr != nil {
		return nil, fmt.Errorf("create invitation use case: save %q: %w", inv.ID(), saveErr)
	}

	return &CreateInvitationResult{Invitation: inv, PlainToken: plainToken}, nil
}
