package match

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/events"
)

// seedMatchInStatus helper: creates and saves a match with the given status.
// Used to set up preconditions for transition tests.
func seedMatchInStatus(t *testing.T, repo *fakeMatchRepository, status entities.MatchStatus) *entities.Match {
	t.Helper()

	match, err := entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          "match-1",
		GroupID:     "group-1",
		Title:       "Test match",
		Venue:       "Venue",
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Status:      status,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("seed: NewMatch failed: %v", err)
	}
	if saveErr := repo.Save(context.Background(), match); saveErr != nil {
		t.Fatalf("seed: Save failed: %v", saveErr)
	}
	return match
}

func TestOpenMatchUseCase_Execute_TransitionsDraftToOpen(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	seedMatchInStatus(t, repo, entities.MatchStatusDraft)
	useCase := NewOpenMatchUseCase(repo)

	match, err := useCase.Execute(context.Background(), "match-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.Status() != entities.MatchStatusOpen {
		t.Errorf("expected status open, got %q", match.Status())
	}
}

func TestMarkTeamsReadyUseCase_Execute_TransitionsOpenToTeamsReady(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	seedMatchInStatus(t, repo, entities.MatchStatusOpen)
	useCase := NewMarkTeamsReadyUseCase(repo, newFakePublisher(), newFakeClock(time.Now()))

	match, err := useCase.Execute(context.Background(), "match-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.Status() != entities.MatchStatusTeamsReady {
		t.Errorf("expected status teams_ready, got %q", match.Status())
	}
}

func TestMarkTeamsReadyUseCase_Execute_PublishesMatchTeamsReadyEvent(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	seedMatchInStatus(t, repo, entities.MatchStatusOpen)
	publisher := newFakePublisher()
	occurredAt := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	useCase := NewMarkTeamsReadyUseCase(repo, publisher, newFakeClock(occurredAt))

	if _, err := useCase.Execute(context.Background(), "match-1"); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(publisher.published) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(publisher.published))
	}
	got, ok := publisher.published[0].(events.MatchTeamsReady)
	if !ok {
		t.Fatalf("expected MatchTeamsReady event, got %T", publisher.published[0])
	}
	if got.MatchID != "match-1" || got.GroupID != "group-1" {
		t.Errorf("event ids mismatch: %+v", got)
	}
	if !got.OccurredAt.Equal(occurredAt) {
		t.Errorf("event timestamp mismatch: got %v want %v", got.OccurredAt, occurredAt)
	}
}

func TestMarkTeamsReadyUseCase_Execute_DoesNotPublish_WhenTransitionFails(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	// Seed in Draft so MarkTeamsReady fails (Open is required).
	seedMatchInStatus(t, repo, entities.MatchStatusDraft)
	publisher := newFakePublisher()
	useCase := NewMarkTeamsReadyUseCase(repo, publisher, newFakeClock(time.Now()))

	if _, err := useCase.Execute(context.Background(), "match-1"); err == nil {
		t.Fatalf("expected error from invalid transition, got nil")
	}
	if len(publisher.published) != 0 {
		t.Errorf("publisher must not be called when transition fails, got %d events", len(publisher.published))
	}
}

func TestStartMatchUseCase_Execute_TransitionsTeamsReadyToInProgress(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	seedMatchInStatus(t, repo, entities.MatchStatusTeamsReady)
	useCase := NewStartMatchUseCase(repo)

	match, err := useCase.Execute(context.Background(), "match-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.Status() != entities.MatchStatusInProgress {
		t.Errorf("expected status in_progress, got %q", match.Status())
	}
}

func TestCompleteMatchUseCase_Execute_TransitionsInProgressToCompleted(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	seedMatchInStatus(t, repo, entities.MatchStatusInProgress)
	useCase := NewCompleteMatchUseCase(repo)

	match, err := useCase.Execute(context.Background(), "match-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.Status() != entities.MatchStatusCompleted {
		t.Errorf("expected status completed, got %q", match.Status())
	}
}

func TestOpenMatchUseCase_Execute_ReturnsErrMatchNotFound_WhenMatchMissing(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	useCase := NewOpenMatchUseCase(repo)

	match, err := useCase.Execute(context.Background(), "nonexistent")

	if match != nil {
		t.Errorf("expected nil match, got %+v", match)
	}
	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestStartMatchUseCase_Execute_ReturnsErrInvalidTransition_WhenMatchIsDraft(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	// A Draft match cannot be started directly — it must go through Open → TeamsReady first.
	seedMatchInStatus(t, repo, entities.MatchStatusDraft)
	useCase := NewStartMatchUseCase(repo)

	match, err := useCase.Execute(context.Background(), "match-1")

	if match != nil {
		t.Errorf("expected nil match, got %+v", match)
	}
	if !errors.Is(err, domainerrors.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}
