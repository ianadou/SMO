package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func newCreateInvitationFixture(t *testing.T, now time.Time) (
	*CreateInvitationUseCase, *fakeInvitationRepository,
) {
	t.Helper()
	repo := newFakeInvitationRepository()
	matchRepo := newFakeMatchRepo()
	playerRepo := newFakePlayerRepo()
	matchRepo.seedMatch(t, "match-1", "group-1")
	playerRepo.seedPlayer(t, "player-1", "group-1")
	uc := NewCreateInvitationUseCase(
		repo, matchRepo, playerRepo,
		newFakeTokenService("plain-token-abc"),
		newFakeIDGenerator("inv-1"),
		newFakeClock(now),
	)
	return uc, repo
}

func TestCreateInvitationUseCase_Execute_ReturnsPlainTokenAndInvitation(t *testing.T) {
	t.Parallel()
	uc, _ := newCreateInvitationFixture(t, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))

	result, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID:  "match-1",
		PlayerID: "player-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlainToken != "plain-token-abc" {
		t.Errorf("expected plain token 'plain-token-abc', got %q", result.PlainToken)
	}
	if result.Invitation.PlayerID() != "player-1" {
		t.Errorf("expected playerID 'player-1', got %q", result.Invitation.PlayerID())
	}
}

func TestCreateInvitationUseCase_Execute_AppliesDefaultValidityWindow_WhenExpiresAtIsZero(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	uc, _ := newCreateInvitationFixture(t, now)

	result, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID: "match-1", PlayerID: "player-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := now.Add(DefaultInvitationValidityDuration)
	if !result.Invitation.ExpiresAt().Equal(expected) {
		t.Errorf("expected expiresAt %v, got %v", expected, result.Invitation.ExpiresAt())
	}
}

func TestCreateInvitationUseCase_Execute_UsesProvidedExpiresAt(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	customExpires := now.Add(24 * time.Hour)
	uc, _ := newCreateInvitationFixture(t, now)

	result, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID:   "match-1",
		PlayerID:  "player-1",
		ExpiresAt: customExpires,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Invitation.ExpiresAt().Equal(customExpires) {
		t.Errorf("expected expiresAt %v, got %v", customExpires, result.Invitation.ExpiresAt())
	}
}

func TestCreateInvitationUseCase_Execute_PropagatesSaveError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	uc, repo := newCreateInvitationFixture(t, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))
	repo.saveErr = repoErr

	_, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID: "match-1", PlayerID: "player-1",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}

func TestCreateInvitationUseCase_Execute_ReturnsErrInvalidID_WhenPlayerIDIsEmpty(t *testing.T) {
	t.Parallel()
	uc, _ := newCreateInvitationFixture(t, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))

	_, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID: "match-1", PlayerID: "",
	})

	if !errors.Is(err, domainerrors.ErrInvalidID) {
		t.Errorf("expected ErrInvalidID, got %v", err)
	}
}

func TestCreateInvitationUseCase_Execute_ReturnsErrMatchNotFound_WhenMatchDoesNotExist(t *testing.T) {
	t.Parallel()
	uc, _ := newCreateInvitationFixture(t, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))

	_, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID: "ghost-match", PlayerID: "player-1",
	})

	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestCreateInvitationUseCase_Execute_ReturnsErrPlayerNotFound_WhenPlayerDoesNotExist(t *testing.T) {
	t.Parallel()
	uc, _ := newCreateInvitationFixture(t, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))

	_, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID: "match-1", PlayerID: "ghost-player",
	})

	if !errors.Is(err, domainerrors.ErrPlayerNotFound) {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestCreateInvitationUseCase_Execute_RejectsPlayerFromOtherGroup(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	repo := newFakeInvitationRepository()
	matchRepo := newFakeMatchRepo()
	playerRepo := newFakePlayerRepo()

	matchRepo.seedMatch(t, "match-1", "group-A")
	playerRepo.seedPlayer(t, "player-X", "group-B")

	uc := NewCreateInvitationUseCase(
		repo, matchRepo, playerRepo,
		newFakeTokenService("tok"),
		newFakeIDGenerator("inv-1"),
		newFakeClock(now),
	)

	// Critical isolation rule: an organizer cannot invite a player from
	// a different group to one of their own matches.
	_, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID:  "match-1",
		PlayerID: "player-X",
	})

	if !errors.Is(err, domainerrors.ErrReferencedEntityNotFound) {
		t.Errorf("expected ErrReferencedEntityNotFound (cross-group), got %v", err)
	}
	if len(repo.invitations) != 0 {
		t.Errorf("invitation must not be persisted on cross-group rejection, found %d", len(repo.invitations))
	}
}

// Sanity check on entities.MaxParticipantsPerMatch — referenced by the
// next use case (AcceptInvitation match-full check) so a regression on
// the constant fails this test loudly here.
func TestMaxParticipantsPerMatch_IsTen(t *testing.T) {
	t.Parallel()
	if entities.MaxParticipantsPerMatch != 10 {
		t.Errorf("expected MaxParticipantsPerMatch=10, got %d", entities.MaxParticipantsPerMatch)
	}
}
