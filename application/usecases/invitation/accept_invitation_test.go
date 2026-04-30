package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// setupAcceptTest creates a repo with an invitation whose hash matches
// the given plain token, and returns the wired use case.
func setupAcceptTest(t *testing.T, plain string, expiresAt time.Time, usedAt *time.Time, clockNow time.Time) *AcceptInvitationUseCase {
	t.Helper()
	repo := newFakeInvitationRepository()
	tokens := newFakeTokenService() // GenerateToken not used here
	hash := tokens.HashToken(plain)

	createdAt := expiresAt.Add(-2 * time.Hour) // must be before expiresAt
	inv, err := entities.NewInvitation("inv-1", "match-1", "p-1", hash, expiresAt, usedAt, createdAt)
	if err != nil {
		t.Fatalf("setup: NewInvitation: %v", err)
	}
	_ = repo.Save(context.Background(), inv)

	return NewAcceptInvitationUseCase(repo, tokens, newFakeClock(clockNow))
}

func TestAcceptInvitationUseCase_Execute_MarksInvitationAsUsed(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	uc := setupAcceptTest(t, "plain-token", expires, nil, now)

	inv, err := uc.Execute(context.Background(), "plain-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inv.IsUsed() {
		t.Error("expected invitation to be used after Accept")
	}
}

func TestAcceptInvitationUseCase_Execute_ReturnsErrInvitationNotFound_WhenTokenDoesNotMatch(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	uc := setupAcceptTest(t, "real-token", expires, nil, now)

	_, err := uc.Execute(context.Background(), "wrong-token")
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestAcceptInvitationUseCase_Execute_ReturnsErrInvitationExpired(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	// Expires in the past relative to now.
	expiresInPast := now.Add(-time.Hour)
	uc := setupAcceptTest(t, "plain-token", expiresInPast, nil, now)

	_, err := uc.Execute(context.Background(), "plain-token")
	if !errors.Is(err, domainerrors.ErrInvitationExpired) {
		t.Errorf("expected ErrInvitationExpired, got %v", err)
	}
}

func TestAcceptInvitationUseCase_Execute_ReturnsErrInvitationAlreadyUsed(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	usedAt := now.Add(-30 * time.Minute) // used earlier
	uc := setupAcceptTest(t, "plain-token", expires, &usedAt, now)

	_, err := uc.Execute(context.Background(), "plain-token")
	if !errors.Is(err, domainerrors.ErrInvitationAlreadyUsed) {
		t.Errorf("expected ErrInvitationAlreadyUsed, got %v", err)
	}
}

func TestAcceptInvitationUseCase_Execute_PropagatesPersistError(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	persistErr := errors.New("disk full")

	tokens := newFakeTokenService()
	hash := tokens.HashToken("plain-token")
	createdAt := expires.Add(-2 * time.Hour)
	inv, err := entities.NewInvitation("inv-1", "match-1", "p-1", hash, expires, nil, createdAt)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	repo := newFakeInvitationRepository()
	_ = repo.Save(context.Background(), inv)
	repo.markAsUsedErr = persistErr

	_, execErr := NewAcceptInvitationUseCase(repo, tokens, newFakeClock(now)).
		Execute(context.Background(), "plain-token")

	if !errors.Is(execErr, persistErr) {
		t.Errorf("expected wrapped persist error, got %v", execErr)
	}
}
