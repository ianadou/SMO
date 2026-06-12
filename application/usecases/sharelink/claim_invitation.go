package sharelink

import (
	"context"
	"fmt"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// ClaimInvitationUseCase binds a roster invitation to the visitor who
// picked their name on the share page. Claiming rotates the invitation
// token: a fresh personal token is minted and returned once, and any
// previously shared personal link for that invitation stops resolving.
type ClaimInvitationUseCase struct {
	links          ports.MatchShareLinkRepository
	invitationRepo ports.InvitationRepository
	matchRepo      ports.MatchRepository
	playerRepo     ports.PlayerRepository
	tokens         ports.InvitationTokenService
	clock          ports.Clock
}

// ClaimInvitationResult bundles the freshly minted personal token with
// the claimed player's display name. Callers must surface the token to
// the visitor exactly once.
type ClaimInvitationResult struct {
	PlainToken string
	PlayerName string
}

// NewClaimInvitationUseCase builds the use case.
func NewClaimInvitationUseCase(
	links ports.MatchShareLinkRepository,
	invitationRepo ports.InvitationRepository,
	matchRepo ports.MatchRepository,
	playerRepo ports.PlayerRepository,
	tokens ports.InvitationTokenService,
	clock ports.Clock,
) *ClaimInvitationUseCase {
	return &ClaimInvitationUseCase{
		links:          links,
		invitationRepo: invitationRepo,
		matchRepo:      matchRepo,
		playerRepo:     playerRepo,
		tokens:         tokens,
		clock:          clock,
	}
}

// Execute claims the invitation of the given player on the match the
// share token designates. Returns ErrShareLinkNotFound /
// ErrShareLinkInactive for a dead link, ErrInvitationLocked once the
// match locks attendance, ErrInvitationNotFound when the player has no
// invitation on this match, and ErrInvitationAlreadyClaimed /
// ErrInvitationExpired from the entity guard.
func (uc *ClaimInvitationUseCase) Execute(
	ctx context.Context,
	shareToken string,
	playerID entities.PlayerID,
) (*ClaimInvitationResult, error) {
	if playerID == "" {
		return nil, fmt.Errorf("claim invitation use case: %w", domainerrors.ErrInvalidID)
	}

	now := uc.clock.Now()

	link, err := findActiveShareLink(ctx, uc.links, uc.tokens, shareToken, now)
	if err != nil {
		return nil, fmt.Errorf("claim invitation use case: %w", err)
	}

	match, err := uc.matchRepo.FindByID(ctx, link.MatchID())
	if err != nil {
		return nil, fmt.Errorf("claim invitation use case: find match: %w", err)
	}

	if match.AttendanceLocked() {
		return nil, fmt.Errorf("claim invitation use case: %w", domainerrors.ErrInvitationLocked)
	}

	inv, err := uc.findMatchInvitationForPlayer(ctx, match.ID(), playerID)
	if err != nil {
		return nil, fmt.Errorf("claim invitation use case: %w", err)
	}

	plainToken, err := uc.tokens.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("claim invitation use case: generate token: %w", err)
	}

	if claimErr := inv.Claim(uc.tokens.HashToken(plainToken), now); claimErr != nil {
		return nil, fmt.Errorf("claim invitation use case: claim: %w", claimErr)
	}

	// The repository claim is conditional (stored row still unclaimed
	// and pending), so the loser of a concurrent claim race surfaces
	// ErrInvitationAlreadyClaimed even though the entity guard above
	// passed on its stale copy.
	if claimErr := uc.invitationRepo.Claim(ctx, inv); claimErr != nil {
		return nil, fmt.Errorf("claim invitation use case: persist %q: %w", inv.ID(), claimErr)
	}

	player, err := uc.playerRepo.FindByID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("claim invitation use case: find player: %w", err)
	}

	return &ClaimInvitationResult{PlainToken: plainToken, PlayerName: player.Name()}, nil
}

// findMatchInvitationForPlayer scans the match's invitations for the
// one issued to the given player, or reports ErrInvitationNotFound.
func (uc *ClaimInvitationUseCase) findMatchInvitationForPlayer(
	ctx context.Context,
	matchID entities.MatchID,
	playerID entities.PlayerID,
) (*entities.Invitation, error) {
	invitations, err := uc.invitationRepo.ListByMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}

	for _, inv := range invitations {
		if inv.PlayerID() == playerID {
			return inv, nil
		}
	}
	return nil, fmt.Errorf("no invitation for player %q on match %q: %w",
		playerID, matchID, domainerrors.ErrInvitationNotFound)
}
