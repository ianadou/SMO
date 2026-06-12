package sharelink

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

type claimFixture struct {
	uc          *ClaimInvitationUseCase
	links       *fakeShareLinkRepository
	invitations *fakeInvitationRepository
	matches     *fakeMatchRepository
	players     *fakePlayerRepository
	tokens      *fakeTokenService
}

// newClaimFixture wires the use case around match-1 (group-1), an
// active link resolving "share-token", and player-1 "Alice". The next
// minted personal token is "new-personal-token". Tests seed the
// invitation themselves so its claimability stays visible in the test
// body.
func newClaimFixture(t *testing.T, now time.Time) *claimFixture {
	t.Helper()
	links := newFakeShareLinkRepository(now)
	players := newFakePlayerRepository()
	invitations := newFakeInvitationRepository(players)
	matches := newFakeMatchRepository()
	tokens := newFakeTokenService("new-personal-token")

	matches.seedMatch(t, "match-1", "group-1")
	players.seedPlayer(t, "player-1", "Alice")
	links.seedLink(t, "link-1",
		tokens.HashToken("share-token"), now.Add(48*time.Hour), nil, now.Add(-time.Hour))

	uc := NewClaimInvitationUseCase(
		links, invitations, matches, players, tokens, newFakeClock(now),
	)
	return &claimFixture{
		uc: uc, links: links, invitations: invitations,
		matches: matches, players: players, tokens: tokens,
	}
}

func (f *claimFixture) seedClaimableInvitation(t *testing.T, now time.Time) {
	t.Helper()
	f.invitations.seedInvitation(t, "inv-1", "player-1", "old-personal-hash",
		now.Add(24*time.Hour), entities.InvitationResponsePending, nil, nil, now.Add(-time.Hour))
}

func TestClaimInvitationUseCase_ReturnsRotatedTokenAndPlayerName_WhenInvitationIsClaimable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newClaimFixture(t, now)
	fixture.seedClaimableInvitation(t, now)

	result, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlainToken != "new-personal-token" {
		t.Errorf("expected plain token 'new-personal-token', got %q", result.PlainToken)
	}
	if result.PlayerName != "Alice" {
		t.Errorf("expected player name 'Alice', got %q", result.PlayerName)
	}
	claimed := fixture.invitations.invitations["inv-1"]
	expectedHash := newFakeTokenService().HashToken("new-personal-token")
	if claimed.TokenHash() != expectedHash {
		t.Errorf("expected rotated token hash, got %q", claimed.TokenHash())
	}
	if claimed.ClaimedAt() == nil || !claimed.ClaimedAt().Equal(now) {
		t.Errorf("expected claimedAt %v, got %v", now, claimed.ClaimedAt())
	}
}

func TestClaimInvitationUseCase_ReturnsErrInvalidID_WhenPlayerIDIsEmpty(t *testing.T) {
	t.Parallel()
	fixture := newClaimFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := fixture.uc.Execute(context.Background(), "share-token", "")

	if !errors.Is(err, domainerrors.ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
}

func TestClaimInvitationUseCase_ReturnsErrShareLinkNotFound_WhenTokenIsUnknown(t *testing.T) {
	t.Parallel()
	fixture := newClaimFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := fixture.uc.Execute(context.Background(), "ghost-token", "player-1")

	if !errors.Is(err, domainerrors.ErrShareLinkNotFound) {
		t.Errorf("expected ErrShareLinkNotFound, got %v", err)
	}
}

func TestClaimInvitationUseCase_ReturnsErrShareLinkInactive_WhenLinkIsRevoked(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := now.Add(-10 * time.Minute)
	fixture := newClaimFixture(t, now)
	tokens := newFakeTokenService()
	fixture.links.seedLink(t, "link-revoked",
		tokens.HashToken("revoked-token"), now.Add(48*time.Hour), &revokedAt, now.Add(-time.Hour))

	_, err := fixture.uc.Execute(context.Background(), "revoked-token", "player-1")

	if !errors.Is(err, domainerrors.ErrShareLinkInactive) {
		t.Errorf("expected ErrShareLinkInactive, got %v", err)
	}
}

func TestClaimInvitationUseCase_ReturnsErrInvitationLocked_WhenAttendanceIsLocked(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newClaimFixture(t, now)
	fixture.matches.seedMatchWithStatus(t, entities.MatchStatusTeamsReady)
	fixture.seedClaimableInvitation(t, now)

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, domainerrors.ErrInvitationLocked) {
		t.Errorf("expected ErrInvitationLocked, got %v", err)
	}
}

func TestClaimInvitationUseCase_ReturnsErrInvitationNotFound_WhenPlayerHasNoInvitationOnMatch(t *testing.T) {
	t.Parallel()
	fixture := newClaimFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestClaimInvitationUseCase_ReturnsErrInvitationAlreadyClaimed_WhenInvitationWasClaimed(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	claimedAt := now.Add(-30 * time.Minute)
	fixture := newClaimFixture(t, now)
	fixture.invitations.seedInvitation(t, "inv-1", "player-1", "old-personal-hash",
		now.Add(24*time.Hour), entities.InvitationResponsePending, nil, &claimedAt, now.Add(-time.Hour))

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, domainerrors.ErrInvitationAlreadyClaimed) {
		t.Errorf("expected ErrInvitationAlreadyClaimed, got %v", err)
	}
}

func TestClaimInvitationUseCase_ReturnsErrInvitationAlreadyClaimed_WhenOwnerAlreadyResponded(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	respondedAt := now.Add(-30 * time.Minute)
	fixture := newClaimFixture(t, now)
	fixture.invitations.seedInvitation(t, "inv-1", "player-1", "old-personal-hash",
		now.Add(24*time.Hour), entities.InvitationResponseYes, &respondedAt, nil, now.Add(-time.Hour))

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, domainerrors.ErrInvitationAlreadyClaimed) {
		t.Errorf("expected ErrInvitationAlreadyClaimed, got %v", err)
	}
}

func TestClaimInvitationUseCase_PropagatesClaimPersistenceError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newClaimFixture(t, now)
	fixture.seedClaimableInvitation(t, now)
	fixture.invitations.claimErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestClaimInvitationUseCase_PropagatesListInvitationsError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db down")
	fixture := newClaimFixture(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC))
	fixture.invitations.listByMatchErr = repoErr

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestClaimInvitationUseCase_ReturnsErrInvitationExpired_WhenInvitationExpired(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newClaimFixture(t, now)
	fixture.invitations.seedInvitation(t, "inv-1", "player-1", "old-personal-hash",
		now.Add(-time.Hour), entities.InvitationResponsePending, nil, nil, now.Add(-2*time.Hour))

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, domainerrors.ErrInvitationExpired) {
		t.Errorf("expected ErrInvitationExpired, got %v", err)
	}
}

func TestClaimInvitationUseCase_PropagatesError_WhenMatchLookupFails(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newClaimFixture(t, now)
	fixture.seedClaimableInvitation(t, now)
	delete(fixture.matches.matches, "match-1")

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestClaimInvitationUseCase_PropagatesError_WhenTokenGenerationFails(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newClaimFixture(t, now)
	fixture.seedClaimableInvitation(t, now)
	entropyErr := errors.New("entropy exhausted")
	fixture.tokens.generateErr = entropyErr

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, entropyErr) {
		t.Errorf("expected wrapped token generation error, got %v", err)
	}
}

func TestClaimInvitationUseCase_PropagatesError_WhenPlayerLookupFailsAfterClaim(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	fixture := newClaimFixture(t, now)
	fixture.seedClaimableInvitation(t, now)
	lookupErr := errors.New("connection reset")
	fixture.players.findErr = lookupErr

	_, err := fixture.uc.Execute(context.Background(), "share-token", "player-1")

	if !errors.Is(err, lookupErr) {
		t.Errorf("expected wrapped player lookup error, got %v", err)
	}
}
