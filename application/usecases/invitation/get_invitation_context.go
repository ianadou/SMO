package invitation

import (
	"context"
	"fmt"
	"time"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/domain/ports"
)

// PageContext is the assembled, display-oriented view a player needs to
// render the invitation page: who is organizing, which match, where and
// when, how many seats, who has already confirmed, and the invitee's own
// current standing (response, lock, expiry).
//
// ConfirmedNames holds the full names of confirmed participants. It is
// intentionally kept at the application boundary: the HTTP layer derives
// initials from it and never exposes the names to the (unauthenticated)
// token bearer.
type PageContext struct {
	OrganizerName   string
	GroupName       string
	MatchTitle      string
	Venue           string
	ScheduledAt     time.Time
	MatchStatus     entities.MatchStatus
	MaxParticipants int
	ConfirmedNames  []string
	Response        entities.InvitationResponse
	ExpiresAt       time.Time
	Locked          bool
	Expired         bool
}

// GetInvitationContextUseCase resolves a plain invitation token into the
// full context needed to render the invitation page. It is a pure read:
// it never mutates the invitation, so an expired or locked invitation
// still returns its context (the state is reported, not rejected).
type GetInvitationContextUseCase struct {
	repo          ports.InvitationRepository
	matchRepo     ports.MatchRepository
	groupRepo     ports.GroupRepository
	organizerRepo ports.OrganizerRepository
	tokens        ports.InvitationTokenService
	clock         ports.Clock
}

// NewGetInvitationContextUseCase builds the use case.
func NewGetInvitationContextUseCase(
	repo ports.InvitationRepository,
	matchRepo ports.MatchRepository,
	groupRepo ports.GroupRepository,
	organizerRepo ports.OrganizerRepository,
	tokens ports.InvitationTokenService,
	clock ports.Clock,
) *GetInvitationContextUseCase {
	return &GetInvitationContextUseCase{
		repo:          repo,
		matchRepo:     matchRepo,
		groupRepo:     groupRepo,
		organizerRepo: organizerRepo,
		tokens:        tokens,
		clock:         clock,
	}
}

// Execute resolves the plain token and assembles the context. Returns
// ErrInvitationNotFound if no invitation matches the hash; any other
// error is an internal data-integrity failure (a dangling match, group
// or organizer reference) and is wrapped for the caller.
func (uc *GetInvitationContextUseCase) Execute(
	ctx context.Context,
	plainToken string,
) (*PageContext, error) {
	hash := uc.tokens.HashToken(plainToken)

	inv, err := uc.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get invitation context use case: find by hash: %w", err)
	}

	match, err := uc.matchRepo.FindByID(ctx, inv.MatchID())
	if err != nil {
		return nil, fmt.Errorf("get invitation context use case: find match: %w", err)
	}

	group, err := uc.groupRepo.FindByID(ctx, match.GroupID())
	if err != nil {
		return nil, fmt.Errorf("get invitation context use case: find group: %w", err)
	}

	organizer, err := uc.organizerRepo.FindByID(ctx, group.OrganizerID())
	if err != nil {
		return nil, fmt.Errorf("get invitation context use case: find organizer: %w", err)
	}

	participants, err := uc.repo.ListConfirmedParticipants(ctx, inv.MatchID())
	if err != nil {
		return nil, fmt.Errorf("get invitation context use case: list confirmed: %w", err)
	}

	confirmedNames := make([]string, 0, len(participants))
	for _, p := range participants {
		confirmedNames = append(confirmedNames, p.PlayerName)
	}

	return &PageContext{
		OrganizerName:   organizer.DisplayName(),
		GroupName:       group.Name(),
		MatchTitle:      match.Title(),
		Venue:           match.Venue(),
		ScheduledAt:     match.ScheduledAt(),
		MatchStatus:     match.Status(),
		MaxParticipants: entities.MaxParticipantsPerMatch,
		ConfirmedNames:  confirmedNames,
		Response:        inv.Response(),
		ExpiresAt:       inv.ExpiresAt(),
		Locked:          match.AttendanceLocked(),
		Expired:         inv.IsExpired(uc.clock.Now()),
	}, nil
}
