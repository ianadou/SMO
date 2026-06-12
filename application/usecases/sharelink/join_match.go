package sharelink

import (
	"context"
	"fmt"
	"strings"

	"github.com/ianadou/smo/application/usecases/invitation"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// JoinMatchUseCase lets a visitor who is not on the roster add
// themselves through the share link. The submitted name is resolved
// against the group's existing players (trimmed, case-insensitive) so a
// returning player keeps their ranking instead of being silently
// duplicated; a genuinely unknown name becomes a new group player with
// the default ranking. Either way a personal invitation is minted and
// its plain token returned once.
//
// Accepted v1 quirk: a genuinely new person homonymous with a past
// group member inherits that player's ranking (see the design doc).
type JoinMatchUseCase struct {
	links          ports.MatchShareLinkRepository
	invitationRepo ports.InvitationRepository
	matchRepo      ports.MatchRepository
	playerRepo     ports.PlayerRepository
	tokens         ports.InvitationTokenService
	idGen          ports.IDGenerator
	clock          ports.Clock
}

// JoinMatchInput is the input of Execute. Both fields come from the
// public share page: the token from the URL, the name from the visitor.
type JoinMatchInput struct {
	ShareToken string
	PlayerName string
}

// JoinMatchResult bundles the created invitation with its plain token
// and the resolved player name (the group's canonical spelling when the
// visitor matched an existing player). Callers must surface the token
// to the visitor exactly once.
type JoinMatchResult struct {
	Invitation *entities.Invitation
	PlainToken string
	PlayerName string
}

// NewJoinMatchUseCase builds the use case.
func NewJoinMatchUseCase(
	links ports.MatchShareLinkRepository,
	invitationRepo ports.InvitationRepository,
	matchRepo ports.MatchRepository,
	playerRepo ports.PlayerRepository,
	tokens ports.InvitationTokenService,
	idGen ports.IDGenerator,
	clock ports.Clock,
) *JoinMatchUseCase {
	return &JoinMatchUseCase{
		links:          links,
		invitationRepo: invitationRepo,
		matchRepo:      matchRepo,
		playerRepo:     playerRepo,
		tokens:         tokens,
		idGen:          idGen,
		clock:          clock,
	}
}

// Execute adds the named visitor to the match. Returns
// ErrShareLinkNotFound / ErrShareLinkInactive for a dead link,
// ErrInvitationLocked once the match locks attendance, ErrInvalidName
// for a blank name, and ErrPlayerAlreadyInvited when the name resolves
// to a player who already has an invitation on this match (the right
// move is to claim that name from the roster instead).
func (uc *JoinMatchUseCase) Execute(
	ctx context.Context,
	input JoinMatchInput,
) (*JoinMatchResult, error) {
	name := strings.TrimSpace(input.PlayerName)
	if name == "" {
		return nil, fmt.Errorf("join match use case: %w", domainerrors.ErrInvalidName)
	}

	now := uc.clock.Now()

	link, err := findActiveShareLink(ctx, uc.links, uc.tokens, input.ShareToken, now)
	if err != nil {
		return nil, fmt.Errorf("join match use case: %w", err)
	}

	match, err := uc.matchRepo.FindByID(ctx, link.MatchID())
	if err != nil {
		return nil, fmt.Errorf("join match use case: find match: %w", err)
	}

	if match.AttendanceLocked() {
		return nil, fmt.Errorf("join match use case: %w", domainerrors.ErrInvitationLocked)
	}

	player, err := uc.resolvePlayer(ctx, match, name)
	if err != nil {
		return nil, fmt.Errorf("join match use case: %w", err)
	}

	plainToken, err := uc.tokens.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("join match use case: generate token: %w", err)
	}

	// Born claimed: the joiner owns the freshly minted token, so the
	// roster must lock this name immediately — otherwise anyone with
	// the share link could claim it and rotate the token away.
	inv, err := entities.NewInvitation(
		entities.InvitationID(uc.idGen.Generate()),
		match.ID(),
		player.ID(),
		uc.tokens.HashToken(plainToken),
		now.Add(invitation.DefaultInvitationValidityDuration),
		entities.InvitationResponsePending,
		nil,
		&now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("join match use case: build invitation: %w", err)
	}

	if saveErr := uc.invitationRepo.Save(ctx, inv); saveErr != nil {
		return nil, fmt.Errorf("join match use case: save invitation %q: %w", inv.ID(), saveErr)
	}

	return &JoinMatchResult{Invitation: inv, PlainToken: plainToken, PlayerName: player.Name()}, nil
}

// resolvePlayer maps the submitted name to the player the invitation
// will be issued for: an existing group player when the trimmed name
// matches case-insensitively (rejected with ErrPlayerAlreadyInvited if
// that player is already on the match's roster), or a brand new player
// with the default ranking otherwise.
func (uc *JoinMatchUseCase) resolvePlayer(
	ctx context.Context,
	match *entities.Match,
	name string,
) (*entities.Player, error) {
	groupPlayers, err := uc.playerRepo.ListByGroup(ctx, match.GroupID())
	if err != nil {
		return nil, fmt.Errorf("list group players: %w", err)
	}

	for _, existing := range groupPlayers {
		if !strings.EqualFold(existing.Name(), name) {
			continue
		}

		invited, invitedErr := uc.playerIsInvited(ctx, match.ID(), existing.ID())
		if invitedErr != nil {
			return nil, invitedErr
		}
		if invited {
			return nil, fmt.Errorf("player %q already invited to match %q: %w",
				existing.ID(), match.ID(), domainerrors.ErrPlayerAlreadyInvited)
		}
		return existing, nil
	}

	player, err := entities.NewPlayer(
		entities.PlayerID(uc.idGen.Generate()),
		match.GroupID(),
		name,
		entities.DefaultPlayerRanking(),
	)
	if err != nil {
		return nil, fmt.Errorf("build player: %w", err)
	}

	if saveErr := uc.playerRepo.Save(ctx, player); saveErr != nil {
		return nil, fmt.Errorf("save player %q: %w", player.ID(), saveErr)
	}
	return player, nil
}

// playerIsInvited reports whether the player already has an invitation
// (any state) on the match.
func (uc *JoinMatchUseCase) playerIsInvited(
	ctx context.Context,
	matchID entities.MatchID,
	playerID entities.PlayerID,
) (bool, error) {
	invitations, err := uc.invitationRepo.ListByMatch(ctx, matchID)
	if err != nil {
		return false, fmt.Errorf("list invitations: %w", err)
	}

	for _, inv := range invitations {
		if inv.PlayerID() == playerID {
			return true, nil
		}
	}
	return false, nil
}
