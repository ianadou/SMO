package sharelink

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type contextFixture struct {
	uc          *GetShareLinkContextUseCase
	links       *fakeShareLinkRepository
	invitations *fakeInvitationRepository
	matches     *fakeMatchRepository
	players     *fakePlayerRepository
	groups      *fakeGroupRepository
	organizers  *fakeOrganizerRepository
}

// newContextFixture wires the use case around match-1 (group-1, owned
// by org-1 "Karim") and an active link resolving the plain token
// "share-token". Tests seed players and invitations themselves so the
// data under assertion stays visible in the test body.
func newContextFixture(t *testing.T, now time.Time) *contextFixture {
	t.Helper()
	links := newFakeShareLinkRepository(now)
	players := newFakePlayerRepository()
	invitations := newFakeInvitationRepository(players)
	matches := newFakeMatchRepository()
	groups := newFakeGroupRepository()
	organizers := newFakeOrganizerRepository()
	tokens := newFakeTokenService()

	matches.seedMatch(t, "match-1", "group-1")
	groups.seedGroup(t)
	organizers.seedOrganizer(t, "org-1", "Karim")
	links.seedLink(t, "link-1",
		tokens.HashToken("share-token"), now.Add(48*time.Hour), nil, now.Add(-time.Hour))

	uc := NewGetShareLinkContextUseCase(
		links, invitations, matches, groups, organizers, players,
		tokens, newFakeClock(now),
	)
	return &contextFixture{
		uc: uc, links: links, invitations: invitations,
		matches: matches, players: players,
		groups: groups, organizers: organizers,
	}
}

func TestGetShareLinkContextUseCase_ReturnsMatchAndGroupContext(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	respondedAt := now.Add(-time.Hour)
	fixture := newContextFixture(t, now)
	fixture.players.seedPlayer(t, "player-1", "Alice")
	fixture.invitations.seedInvitation(t, "inv-1", "player-1", "hash-1",
		now.Add(24*time.Hour), entities.InvitationResponseYes, &respondedAt, nil, now.Add(-2*time.Hour))

	pageContext, err := fixture.uc.Execute(context.Background(), "share-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pageContext.MatchID != "match-1" {
		t.Errorf("expected match id 'match-1', got %q", pageContext.MatchID)
	}
	if pageContext.OrganizerName != "Karim" {
		t.Errorf("expected organizer 'Karim', got %q", pageContext.OrganizerName)
	}
	if pageContext.GroupName != "Sunday League" {
		t.Errorf("expected group 'Sunday League', got %q", pageContext.GroupName)
	}
	if pageContext.MatchTitle != "Match" || pageContext.Venue != "Venue" {
		t.Errorf("expected match 'Match' at 'Venue', got %q at %q",
			pageContext.MatchTitle, pageContext.Venue)
	}
	if pageContext.MaxParticipants != entities.MaxParticipantsPerMatch {
		t.Errorf("expected max participants %d, got %d",
			entities.MaxParticipantsPerMatch, pageContext.MaxParticipants)
	}
	if !reflect.DeepEqual(pageContext.ConfirmedNames, []string{"Alice"}) {
		t.Errorf("expected confirmed names [Alice], got %v", pageContext.ConfirmedNames)
	}
	if pageContext.Locked {
		t.Error("expected attendance not locked for an open match")
	}
}

func TestGetShareLinkContextUseCase_MapsRosterStates(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	respondedAt := now.Add(-time.Hour)
	claimedAt := now.Add(-30 * time.Minute)
	fixture := newContextFixture(t, now)
	fixture.players.seedPlayer(t, "player-1", "Alice")
	fixture.players.seedPlayer(t, "player-2", "Bruno")
	fixture.players.seedPlayer(t, "player-3", "Chloe")
	fixture.players.seedPlayer(t, "player-4", "Dora")
	expiresAt := now.Add(24 * time.Hour)
	createdAt := now.Add(-2 * time.Hour)
	fixture.invitations.seedInvitation(t, "inv-1", "player-1", "hash-1",
		expiresAt, entities.InvitationResponsePending, nil, nil, createdAt)
	fixture.invitations.seedInvitation(t, "inv-2", "player-2", "hash-2",
		expiresAt, entities.InvitationResponsePending, nil, &claimedAt, createdAt)
	fixture.invitations.seedInvitation(t, "inv-3", "player-3", "hash-3",
		expiresAt, entities.InvitationResponseYes, &respondedAt, nil, createdAt)
	fixture.invitations.seedInvitation(t, "inv-4", "player-4", "hash-4",
		expiresAt, entities.InvitationResponseNo, &respondedAt, nil, createdAt)

	pageContext, err := fixture.uc.Execute(context.Background(), "share-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []RosterEntry{
		{PlayerID: "player-1", PlayerName: "Alice", State: RosterStateClaimable},
		{PlayerID: "player-2", PlayerName: "Bruno", State: RosterStateClaimed},
		{PlayerID: "player-3", PlayerName: "Chloe", State: RosterStateResponded},
		{PlayerID: "player-4", PlayerName: "Dora", State: RosterStateResponded},
	}
	if !reflect.DeepEqual(pageContext.Roster, expected) {
		t.Errorf("expected roster %v, got %v", expected, pageContext.Roster)
	}
}

func TestGetShareLinkContextUseCase_ReportsLocked_WhenMatchIsTeamsReady(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newContextFixture(t, now)
	fixture.matches.seedMatchWithStatus(t, entities.MatchStatusTeamsReady)

	pageContext, err := fixture.uc.Execute(context.Background(), "share-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pageContext.Locked {
		t.Error("expected attendance locked from teams_ready onward")
	}
}

func TestGetShareLinkContextUseCase_ReturnsErrShareLinkNotFound_WhenTokenIsUnknown(t *testing.T) {
	t.Parallel()
	fixture := newContextFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := fixture.uc.Execute(context.Background(), "ghost-token")

	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_ReturnsErrShareLinkInactive_WhenLinkIsRevoked(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := now.Add(-10 * time.Minute)
	fixture := newContextFixture(t, now)
	tokens := newFakeTokenService()
	fixture.links.seedLink(t, "link-revoked",
		tokens.HashToken("revoked-token"), now.Add(48*time.Hour), &revokedAt, now.Add(-time.Hour))

	_, err := fixture.uc.Execute(context.Background(), "revoked-token")

	if !errors.Is(err, domainerrors.ErrShareLinkInactive) {
		t.Errorf("expected ErrShareLinkInactive, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_ReturnsErrShareLinkInactive_WhenLinkIsExpired(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newContextFixture(t, now)
	tokens := newFakeTokenService()
	fixture.links.seedLink(t, "link-expired",
		tokens.HashToken("expired-token"), now.Add(-24*time.Hour), nil, now.Add(-72*time.Hour))

	_, err := fixture.uc.Execute(context.Background(), "expired-token")

	if !errors.Is(err, domainerrors.ErrShareLinkInactive) {
		t.Errorf("expected ErrShareLinkInactive, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_PropagatesListConfirmedError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db down")
	fixture := newContextFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	fixture.invitations.listConfirmedErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), "share-token")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_PropagatesListInvitationsError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db down")
	fixture := newContextFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	fixture.invitations.listByMatchErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), "share-token")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_ReturnsErrPlayerNotFound_WhenInvitationReferencesUnknownPlayer(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newContextFixture(t, now)
	fixture.invitations.seedInvitation(t, "inv-1", "ghost-player", "hash-1",
		now.Add(24*time.Hour), entities.InvitationResponsePending, nil, nil, now.Add(-time.Hour))

	_, err := fixture.uc.Execute(context.Background(), "share-token")

	if !errors.Is(err, domainerrors.ErrPlayerNotFound) {
		t.Errorf("expected ErrPlayerNotFound on dangling reference, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_PropagatesError_WhenMatchLookupFails(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newContextFixture(t, now)
	delete(fixture.matches.matches, "match-1")

	_, err := fixture.uc.Execute(context.Background(), "share-token")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_PropagatesError_WhenGroupLookupFails(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newContextFixture(t, now)
	delete(fixture.groups.groups, "group-1")

	_, err := fixture.uc.Execute(context.Background(), "share-token")

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_PropagatesError_WhenOrganizerLookupFails(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newContextFixture(t, now)
	delete(fixture.organizers.organizers, "org-1")

	_, err := fixture.uc.Execute(context.Background(), "share-token")

	if !errors.Is(err, domainerrors.ErrOrganizerNotFound) {
		t.Errorf("expected ErrOrganizerNotFound, got %v", err)
	}
}

func TestGetShareLinkContextUseCase_PropagatesError_WhenGroupPlayersLookupFails(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newContextFixture(t, now)
	repoErr := errors.New("db down")
	fixture.players.listErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), "share-token")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped players lookup error, got %v", err)
	}
}
