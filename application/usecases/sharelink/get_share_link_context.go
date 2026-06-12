package sharelink

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/ports"
)

// RosterEntryState tells the share page how to render a roster name:
// tappable, or locked with the reason it is taken.
type RosterEntryState string

const (
	// RosterStateClaimable means the invitation is untouched: pending
	// and never claimed, so the visitor can take this name.
	RosterStateClaimable RosterEntryState = "claimable"

	// RosterStateClaimed means someone already claimed the invitation
	// through the share link but has not answered yet.
	RosterStateClaimed RosterEntryState = "claimed"

	// RosterStateResponded means the invitation's owner already settled
	// their answer (yes or no), through either link.
	RosterStateResponded RosterEntryState = "responded"
)

// RosterEntry is one selectable name on the share page.
type RosterEntry struct {
	PlayerID   entities.PlayerID
	PlayerName string
	State      RosterEntryState
}

// PageContext is the assembled, display-oriented view a visitor needs
// to render the share page: who is organizing, which match, where and
// when, who has already confirmed, and the roster of invited names with
// their claimability.
//
// ConfirmedNames holds the full names of confirmed participants. It is
// intentionally kept at the application boundary: the HTTP layer derives
// initials from it, like the invitation page does.
//
// MatchID is exposed so the join page can key its local token stash per
// match and recognize a returning visitor.
type PageContext struct {
	MatchID         entities.MatchID
	OrganizerName   string
	GroupName       string
	MatchTitle      string
	Venue           string
	ScheduledAt     time.Time
	MatchStatus     entities.MatchStatus
	MaxParticipants int
	ConfirmedNames  []string
	Locked          bool
	Roster          []RosterEntry
}

// GetShareLinkContextUseCase resolves a plain share token into the full
// context needed to render the share page. It is a pure read: it never
// mutates the link or the invitations.
type GetShareLinkContextUseCase struct {
	links          ports.MatchShareLinkRepository
	invitationRepo ports.InvitationRepository
	matchRepo      ports.MatchRepository
	groupRepo      ports.GroupRepository
	organizerRepo  ports.OrganizerRepository
	playerRepo     ports.PlayerRepository
	tokens         ports.InvitationTokenService
	clock          ports.Clock
}

// NewGetShareLinkContextUseCase builds the use case.
func NewGetShareLinkContextUseCase(
	links ports.MatchShareLinkRepository,
	invitationRepo ports.InvitationRepository,
	matchRepo ports.MatchRepository,
	groupRepo ports.GroupRepository,
	organizerRepo ports.OrganizerRepository,
	playerRepo ports.PlayerRepository,
	tokens ports.InvitationTokenService,
	clock ports.Clock,
) *GetShareLinkContextUseCase {
	return &GetShareLinkContextUseCase{
		links:          links,
		invitationRepo: invitationRepo,
		matchRepo:      matchRepo,
		groupRepo:      groupRepo,
		organizerRepo:  organizerRepo,
		playerRepo:     playerRepo,
		tokens:         tokens,
		clock:          clock,
	}
}

// Execute resolves the plain token and assembles the context. Returns
// ErrShareLinkNotFound for an unknown token and ErrShareLinkInactive
// for a revoked or expired link; any other error is an internal
// data-integrity failure and is wrapped for the caller.
func (uc *GetShareLinkContextUseCase) Execute(
	ctx context.Context,
	plainToken string,
) (*PageContext, error) {
	link, err := findActiveShareLink(ctx, uc.links, uc.tokens, plainToken, uc.clock.Now())
	if err != nil {
		return nil, fmt.Errorf("get share link context use case: %w", err)
	}

	match, err := uc.matchRepo.FindByID(ctx, link.MatchID())
	if err != nil {
		return nil, fmt.Errorf("get share link context use case: find match: %w", err)
	}

	group, err := uc.groupRepo.FindByID(ctx, match.GroupID())
	if err != nil {
		return nil, fmt.Errorf("get share link context use case: find group: %w", err)
	}

	organizer, err := uc.organizerRepo.FindByID(ctx, group.OrganizerID())
	if err != nil {
		return nil, fmt.Errorf("get share link context use case: find organizer: %w", err)
	}

	participants, err := uc.invitationRepo.ListConfirmedParticipants(ctx, match.ID())
	if err != nil {
		return nil, fmt.Errorf("get share link context use case: list confirmed: %w", err)
	}

	confirmedNames := make([]string, 0, len(participants))
	for _, p := range participants {
		confirmedNames = append(confirmedNames, p.PlayerName)
	}

	roster, err := uc.buildRoster(ctx, match)
	if err != nil {
		return nil, fmt.Errorf("get share link context use case: %w", err)
	}

	return &PageContext{
		MatchID:         match.ID(),
		OrganizerName:   organizer.DisplayName(),
		GroupName:       group.Name(),
		MatchTitle:      match.Title(),
		Venue:           match.Venue(),
		ScheduledAt:     match.ScheduledAt(),
		MatchStatus:     match.Status(),
		MaxParticipants: entities.MaxParticipantsPerMatch,
		ConfirmedNames:  confirmedNames,
		Locked:          match.AttendanceLocked(),
		Roster:          roster,
	}, nil
}

// buildRoster projects every invitation of the match into a named entry
// with its claimability. Entries are sorted by player name so the page
// renders a stable list regardless of storage order.
func (uc *GetShareLinkContextUseCase) buildRoster(
	ctx context.Context,
	match *entities.Match,
) ([]RosterEntry, error) {
	invitations, err := uc.invitationRepo.ListByMatch(ctx, match.ID())
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}

	players, err := uc.playerRepo.ListByGroup(ctx, match.GroupID())
	if err != nil {
		return nil, fmt.Errorf("list group players: %w", err)
	}

	nameByID := make(map[entities.PlayerID]string, len(players))
	for _, p := range players {
		nameByID[p.ID()] = p.Name()
	}

	roster := make([]RosterEntry, 0, len(invitations))
	for _, inv := range invitations {
		name, ok := nameByID[inv.PlayerID()]
		if !ok {
			return nil, fmt.Errorf("invitation %q references player %q outside the group: %w",
				inv.ID(), inv.PlayerID(), domainerrors.ErrPlayerNotFound)
		}
		roster = append(roster, RosterEntry{
			PlayerID:   inv.PlayerID(),
			PlayerName: name,
			State:      rosterState(inv),
		})
	}

	sort.Slice(roster, func(i, j int) bool {
		return roster[i].PlayerName < roster[j].PlayerName
	})
	return roster, nil
}

// rosterState derives the claimability of an invitation. A settled
// response wins over the claim mark: an owner who already answered has
// engaged, whether or not the answer came through the share link.
func rosterState(inv *entities.Invitation) RosterEntryState {
	switch {
	case inv.Response() != entities.InvitationResponsePending:
		return RosterStateResponded
	case inv.ClaimedAt() != nil:
		return RosterStateClaimed
	default:
		return RosterStateClaimable
	}
}
