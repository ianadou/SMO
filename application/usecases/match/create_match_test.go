package match

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestCreateMatchUseCase_Execute_ReturnsMatch_WhenInputIsValid(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	scheduledAt := time.Date(2026, 5, 1, 18, 0, 0, 0, time.UTC)

	repo := newFakeMatchRepository()
	idGen := newFakeIDGenerator("match-fixed-id")
	clock := newFakeClock(fixedTime)
	useCase := NewCreateMatchUseCase(repo, idGen, clock)

	input := CreateMatchInput{
		GroupID:     "group-1",
		Title:       "Friday football",
		Venue:       "Stadium A",
		ScheduledAt: scheduledAt,
	}

	match, err := useCase.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.ID() != "match-fixed-id" {
		t.Errorf("expected ID 'match-fixed-id', got %q", match.ID())
	}
	if match.Title() != "Friday football" {
		t.Errorf("expected title 'Friday football', got %q", match.Title())
	}
	if match.Status() != entities.MatchStatusDraft {
		t.Errorf("expected status draft, got %q", match.Status())
	}
	if !match.CreatedAt().Equal(fixedTime) {
		t.Errorf("expected createdAt %v, got %v", fixedTime, match.CreatedAt())
	}
}

func TestCreateMatchUseCase_Execute_PersistsMatch(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	idGen := newFakeIDGenerator("match-1")
	clock := newFakeClock(time.Now())
	useCase := NewCreateMatchUseCase(repo, idGen, clock)

	_, err := useCase.Execute(context.Background(), CreateMatchInput{
		GroupID:     "group-1",
		Title:       "Test match",
		Venue:       "Venue",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stored, findErr := repo.FindByID(context.Background(), "match-1")
	if findErr != nil {
		t.Fatalf("expected match to be persisted, got error: %v", findErr)
	}
	if stored.Title() != "Test match" {
		t.Errorf("expected stored title 'Test match', got %q", stored.Title())
	}
}

func TestCreateMatchUseCase_Execute_ReturnsError_WhenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	idGen := newFakeIDGenerator("match-1")
	clock := newFakeClock(time.Now())
	useCase := NewCreateMatchUseCase(repo, idGen, clock)

	match, err := useCase.Execute(context.Background(), CreateMatchInput{
		GroupID:     "group-1",
		Title:       "",
		Venue:       "Venue",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	})

	if match != nil {
		t.Errorf("expected nil match, got %+v", match)
	}
	if !errors.Is(err, domainerrors.ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

// failingMatchRepository returns an error on Save for error-propagation tests.
type failingMatchRepository struct {
	*fakeMatchRepository
	saveErr error
}

func (r *failingMatchRepository) Save(_ context.Context, _ *entities.Match) error {
	return r.saveErr
}

func TestCreateMatchUseCase_Execute_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database is unreachable")
	repo := &failingMatchRepository{
		fakeMatchRepository: newFakeMatchRepository(),
		saveErr:             expectedErr,
	}
	idGen := newFakeIDGenerator("match-1")
	clock := newFakeClock(time.Now())
	useCase := NewCreateMatchUseCase(repo, idGen, clock)

	match, err := useCase.Execute(context.Background(), CreateMatchInput{
		GroupID:     "group-1",
		Title:       "Valid",
		Venue:       "Venue",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	})

	if match != nil {
		t.Errorf("expected nil match, got %+v", match)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}
