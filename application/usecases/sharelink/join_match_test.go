package sharelink

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type joinFixture struct {
	uc          *JoinMatchUseCase
	links       *fakeShareLinkRepository
	invitations *fakeInvitationRepository
	matches     *fakeMatchRepository
	players     *fakePlayerRepository
	tokens      *fakeTokenService
}

// newJoinFixture wires the use case around match-1 (group-1), an active
// link resolving "share-token", and the existing group player "Marc".
// The next minted personal token is "fresh-token"; the id generator
// serves the given ids in order (player id first when a player must be
// created, invitation id last).
func newJoinFixture(t *testing.T, now time.Time, ids ...string) *joinFixture {
	t.Helper()
	links := newFakeShareLinkRepository(now)
	players := newFakePlayerRepository()
	invitations := newFakeInvitationRepository(players)
	matches := newFakeMatchRepository()
	tokens := newFakeTokenService("fresh-token")

	matches.seedMatch(t, "match-1", "group-1")
	players.seedPlayer(t, "player-1", "Marc")
	links.seedLink(t, "link-1",
		tokens.HashToken("share-token"), now.Add(48*time.Hour), nil, now.Add(-time.Hour))

	uc := NewJoinMatchUseCase(
		links, invitations, matches, players,
		tokens, newFakeIDGenerator(ids...), newFakeClock(now),
	)
	return &joinFixture{
		uc: uc, links: links, invitations: invitations,
		matches: matches, players: players, tokens: tokens,
	}
}

func TestJoinMatchUseCase_InvitesExistingPlayer_WhenNameMatchesIgnoringCaseAndSpaces(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newJoinFixture(t, now, "inv-new")

	result, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "  marc ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlainToken != "fresh-token" {
		t.Errorf("expected plain token 'fresh-token', got %q", result.PlainToken)
	}
	if result.PlayerName != "Marc" {
		t.Errorf("expected canonical name 'Marc', got %q", result.PlayerName)
	}
	if result.Invitation.PlayerID() != "player-1" {
		t.Errorf("expected invitation for existing 'player-1', got %q", result.Invitation.PlayerID())
	}
	if len(fixture.players.players) != 1 {
		t.Errorf("no new player must be created on a name match, found %d", len(fixture.players.players))
	}
	if _, ok := fixture.invitations.invitations["inv-new"]; !ok {
		t.Error("expected invitation to be persisted under id 'inv-new'")
	}
}

func TestJoinMatchUseCase_CreatesPlayerWithDefaultRanking_WhenNameIsUnknown(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newJoinFixture(t, now, "player-new", "inv-new")

	result, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Zoe",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	created, ok := fixture.players.players["player-new"]
	if !ok {
		t.Fatal("expected a new player to be persisted under id 'player-new'")
	}
	if created.Name() != "Zoe" {
		t.Errorf("expected player name 'Zoe', got %q", created.Name())
	}
	if created.GroupID() != "group-1" {
		t.Errorf("expected player in group 'group-1', got %q", created.GroupID())
	}
	if created.Ranking() != entities.DefaultPlayerRanking() {
		t.Errorf("expected default ranking %d, got %d", entities.DefaultPlayerRanking(), created.Ranking())
	}
	if result.Invitation.PlayerID() != "player-new" {
		t.Errorf("expected invitation for 'player-new', got %q", result.Invitation.PlayerID())
	}
}

func TestJoinMatchUseCase_ReturnsErrPlayerAlreadyInvited_WhenNameMatchesInvitedPlayer(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newJoinFixture(t, now)
	fixture.invitations.seedInvitation(t, "inv-1", "player-1", "existing-hash",
		now.Add(24*time.Hour), entities.InvitationResponsePending, nil, nil, now.Add(-time.Hour))

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Marc",
	})

	if !errors.Is(err, domainerrors.ErrPlayerAlreadyInvited) {
		t.Errorf("expected ErrPlayerAlreadyInvited, got %v", err)
	}
	if len(fixture.invitations.invitations) != 1 {
		t.Errorf("no invitation must be created on the conflict, found %d",
			len(fixture.invitations.invitations))
	}
}

func TestJoinMatchUseCase_ReturnsErrInvalidName_WhenNameIsBlank(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "   ",
	})

	if !errors.Is(err, domainerrors.ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestJoinMatchUseCase_ReturnsErrInvitationLocked_WhenAttendanceIsLocked(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newJoinFixture(t, now)
	fixture.matches.seedMatchWithStatus(t, entities.MatchStatusTeamsReady)

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Zoe",
	})

	if !errors.Is(err, domainerrors.ErrInvitationLocked) {
		t.Errorf("expected ErrInvitationLocked, got %v", err)
	}
}

func TestJoinMatchUseCase_ReturnsErrShareLinkNotFound_WhenTokenIsUnknown(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "ghost-token",
		PlayerName: "Zoe",
	})

	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound, got %v", err)
	}
}

func TestJoinMatchUseCase_PropagatesPlayerSaveError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), "player-new")
	fixture.players.saveErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Zoe",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestJoinMatchUseCase_PropagatesInvitationSaveError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), "inv-new")
	fixture.invitations.saveErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Marc",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestJoinMatchUseCase_ReturnsErrShareLinkInactive_WhenLinkIsExpired(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newJoinFixture(t, now)
	tokens := newFakeTokenService()
	fixture.links.seedLink(t, "link-expired",
		tokens.HashToken("expired-token"), now.Add(-24*time.Hour), nil, now.Add(-72*time.Hour))

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "expired-token",
		PlayerName: "Zoe",
	})

	if !errors.Is(err, domainerrors.ErrShareLinkInactive) {
		t.Errorf("expected ErrShareLinkInactive, got %v", err)
	}
}

func TestJoinMatchUseCase_MarksInvitationClaimedAtBirth_SoTheNameIsNotSquattable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newJoinFixture(t, now, "player-new", "inv-new")

	result, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Zoe",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	claimedAt := result.Invitation.ClaimedAt()
	if claimedAt == nil || !claimedAt.Equal(now) {
		t.Errorf("expected invitation claimed at %v so the roster locks the name, got %v", now, claimedAt)
	}
}

func TestJoinMatchUseCase_PropagatesError_WhenMatchLookupFails(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	delete(fixture.matches.matches, "match-1")

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Zoe",
	})

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestJoinMatchUseCase_PropagatesError_WhenTokenGenerationFails(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	entropyErr := errors.New("entropy exhausted")
	fixture.tokens.generateErr = entropyErr

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Marc",
	})

	if !errors.Is(err, entropyErr) {
		t.Errorf("expected wrapped token generation error, got %v", err)
	}
}

func TestJoinMatchUseCase_PropagatesError_WhenGroupPlayersLookupFails(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	repoErr := errors.New("db down")
	fixture.players.listErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Zoe",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped players lookup error, got %v", err)
	}
}

func TestJoinMatchUseCase_PropagatesError_WhenInvitationLookupFails(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	repoErr := errors.New("db down")
	fixture.invitations.listByMatchErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Marc",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped invitations lookup error, got %v", err)
	}
}

func TestJoinMatchUseCase_PropagatesError_WhenGeneratedPlayerIDIsInvalid(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), "")

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Zoe",
	})

	if !errors.Is(err, domainerrors.ErrInvalidID) {
		t.Errorf("expected ErrInvalidID from an empty generated player id, got %v", err)
	}
}

func TestJoinMatchUseCase_PropagatesError_WhenGeneratedInvitationIDIsInvalid(t *testing.T) {
	t.Parallel()
	fixture := newJoinFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), "")

	_, err := fixture.uc.Execute(context.Background(), JoinMatchInput{
		ShareToken: "share-token",
		PlayerName: "Marc",
	})

	if !errors.Is(err, domainerrors.ErrInvalidID) {
		t.Errorf("expected ErrInvalidID from an empty generated invitation id, got %v", err)
	}
}
