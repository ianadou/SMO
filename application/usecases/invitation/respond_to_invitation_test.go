package invitation

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// respondFixture wires a RespondToInvitationUseCase around a single
// invitation whose hash matches plain, plus a match seeded in the given
// status so the lock branch can be exercised.
type respondFixture struct {
	uc   *RespondToInvitationUseCase
	repo *fakeInvitationRepository
}

func newRespondFixture(
	t *testing.T,
	plain string,
	expiresAt time.Time,
	response entities.InvitationResponse,
	respondedAt *time.Time,
	matchStatus entities.MatchStatus,
	clockNow time.Time,
) respondFixture {
	t.Helper()
	repo := newFakeInvitationRepository()
	matchRepo := newFakeMatchRepo()
	matchRepo.seedMatchWithStatus(t, matchStatus)
	tokens := newFakeTokenService()
	hash := tokens.HashToken(plain)

	createdAt := expiresAt.Add(-7 * 24 * time.Hour)
	inv, err := entities.NewInvitation("inv-1", "match-1", "p-1", hash, expiresAt, response, respondedAt, createdAt)
	if err != nil {
		t.Fatalf("setup: NewInvitation: %v", err)
	}
	_ = repo.Save(context.Background(), inv)

	uc := NewRespondToInvitationUseCase(repo, matchRepo, tokens, newFakeClock(clockNow))
	return respondFixture{uc: uc, repo: repo}
}

func TestRespondToInvitationUseCase_Execute_ConfirmsAttendance_WhenAnswerIsYes(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	f := newRespondFixture(t, "tok", expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	inv, err := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inv.IsConfirmed() {
		t.Error("expected invitation to be confirmed after answering yes")
	}
}

func TestRespondToInvitationUseCase_Execute_RecordsDecline_WhenAnswerIsNo(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	f := newRespondFixture(t, "tok", expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	inv, err := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseNo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Response() != entities.InvitationResponseNo {
		t.Errorf("expected response 'no', got %q", inv.Response())
	}
	if inv.IsConfirmed() {
		t.Error("a declined invitation must not be confirmed")
	}
}

func TestRespondToInvitationUseCase_Execute_ChangesYesToNo(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	earlier := now.Add(-time.Hour)
	f := newRespondFixture(t, "tok", expires, entities.InvitationResponseYes, &earlier, entities.MatchStatusOpen, now)

	inv, err := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseNo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Response() != entities.InvitationResponseNo {
		t.Errorf("expected response 'no' after change, got %q", inv.Response())
	}
}

func TestRespondToInvitationUseCase_Execute_ChangesNoToYes(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	earlier := now.Add(-time.Hour)
	f := newRespondFixture(t, "tok", expires, entities.InvitationResponseNo, &earlier, entities.MatchStatusOpen, now)

	inv, err := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inv.IsConfirmed() {
		t.Error("expected invitation to be confirmed after changing no->yes")
	}
}

func TestRespondToInvitationUseCase_Execute_IsIdempotent_WhenSameAnswerRepeated(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	earlier := now.Add(-time.Hour)
	f := newRespondFixture(t, "tok", expires, entities.InvitationResponseYes, &earlier, entities.MatchStatusOpen, now)

	inv, err := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)
	if err != nil {
		t.Fatalf("unexpected error on idempotent yes: %v", err)
	}
	if !inv.IsConfirmed() {
		t.Error("expected invitation to remain confirmed")
	}
	if inv.RespondedAt() == nil || !inv.RespondedAt().Equal(now) {
		t.Errorf("expected RespondedAt refreshed to %v, got %v", now, inv.RespondedAt())
	}
}

func TestRespondToInvitationUseCase_Execute_ReturnsErrInvitationNotFound_WhenTokenDoesNotMatch(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	f := newRespondFixture(t, "real-token", expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	_, err := f.uc.Execute(context.Background(), "wrong-token", entities.InvitationResponseYes)
	if !errors.Is(err, domainerrors.ErrInvitationNotFound) {
		t.Errorf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestRespondToInvitationUseCase_Execute_ReturnsErrInvitationExpired(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresInPast := now.Add(-time.Hour)
	f := newRespondFixture(t, "tok", expiresInPast, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)

	_, err := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)
	if !errors.Is(err, domainerrors.ErrInvitationExpired) {
		t.Errorf("expected ErrInvitationExpired, got %v", err)
	}
}

func TestRespondToInvitationUseCase_Execute_ReturnsErrInvitationLocked_WhenMatchPastOpen(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)

	for _, status := range []entities.MatchStatus{
		entities.MatchStatusTeamsReady,
		entities.MatchStatusInProgress,
		entities.MatchStatusCompleted,
		entities.MatchStatusClosed,
	} {
		t.Run(string(status), func(t *testing.T) {
			t.Parallel()
			f := newRespondFixture(t, "tok", expires, entities.InvitationResponsePending, nil, status, now)

			_, err := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)

			if !errors.Is(err, domainerrors.ErrInvitationLocked) {
				t.Errorf("status %q: expected ErrInvitationLocked, got %v", status, err)
			}
		})
	}
}

func TestRespondToInvitationUseCase_Execute_ReturnsErrMatchFull_WhenMatchAtCapacity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	createdAt := expires.Add(-7 * 24 * time.Hour)

	repo := newFakeInvitationRepository()
	matchRepo := newFakeMatchRepo()
	matchRepo.seedMatchWithStatus(t, entities.MatchStatusOpen)
	tokens := newFakeTokenService()
	hash := tokens.HashToken("tok")

	for i := 0; i < entities.MaxParticipantsPerMatch; i++ {
		respondedAt := createdAt.Add(time.Duration(i) * time.Minute)
		confirmed, err := entities.NewInvitation(
			entities.InvitationID(fmt.Sprintf("inv-confirmed-%d", i)),
			"match-1",
			entities.PlayerID(fmt.Sprintf("player-%d", i)),
			fmt.Sprintf("hash-%d", i),
			expires,
			entities.InvitationResponseYes,
			&respondedAt,
			createdAt,
		)
		if err != nil {
			t.Fatalf("seed confirmed: %v", err)
		}
		_ = repo.Save(context.Background(), confirmed)
	}

	pending, err := entities.NewInvitation(
		"inv-pending", "match-1", "player-late", hash, expires,
		entities.InvitationResponsePending, nil, createdAt,
	)
	if err != nil {
		t.Fatalf("seed pending: %v", err)
	}
	_ = repo.Save(context.Background(), pending)

	uc := NewRespondToInvitationUseCase(repo, matchRepo, tokens, newFakeClock(now))

	_, execErr := uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)

	if !errors.Is(execErr, domainerrors.ErrMatchFull) {
		t.Errorf("expected ErrMatchFull, got %v", execErr)
	}
}

func TestRespondToInvitationUseCase_Execute_AllowsConfirmation_WhenBelowCapacity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	createdAt := expires.Add(-7 * 24 * time.Hour)

	repo := newFakeInvitationRepository()
	matchRepo := newFakeMatchRepo()
	matchRepo.seedMatchWithStatus(t, entities.MatchStatusOpen)
	tokens := newFakeTokenService()
	hash := tokens.HashToken("tok")

	for i := 0; i < entities.MaxParticipantsPerMatch-1; i++ {
		respondedAt := createdAt.Add(time.Duration(i) * time.Minute)
		confirmed, _ := entities.NewInvitation(
			entities.InvitationID(fmt.Sprintf("inv-confirmed-%d", i)),
			"match-1",
			entities.PlayerID(fmt.Sprintf("player-%d", i)),
			fmt.Sprintf("hash-%d", i),
			expires,
			entities.InvitationResponseYes,
			&respondedAt,
			createdAt,
		)
		_ = repo.Save(context.Background(), confirmed)
	}

	pending, _ := entities.NewInvitation(
		"inv-pending", "match-1", "player-last", hash, expires,
		entities.InvitationResponsePending, nil, createdAt,
	)
	_ = repo.Save(context.Background(), pending)

	uc := NewRespondToInvitationUseCase(repo, matchRepo, tokens, newFakeClock(now))

	inv, execErr := uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)

	if execErr != nil {
		t.Fatalf("expected no error at capacity-1, got %v", execErr)
	}
	if !inv.IsConfirmed() {
		t.Errorf("expected last-slot invitation to be confirmed")
	}
}

func TestRespondToInvitationUseCase_Execute_PropagatesPersistError(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expires := now.Add(24 * time.Hour)
	persistErr := errors.New("disk full")

	f := newRespondFixture(t, "tok", expires, entities.InvitationResponsePending, nil, entities.MatchStatusOpen, now)
	f.repo.respondErr = persistErr

	_, execErr := f.uc.Execute(context.Background(), "tok", entities.InvitationResponseYes)

	if !errors.Is(execErr, persistErr) {
		t.Errorf("expected wrapped persist error, got %v", execErr)
	}
}
