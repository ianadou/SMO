package mappers

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

func TestMatchToDomain_ReturnsEntity_WhenRowIsValid(t *testing.T) {
	t.Parallel()

	scheduledAt := time.Date(2026, 5, 1, 18, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 11, 11, 0, 0, 0, time.UTC)
	row := generated.Matches{
		ID:          "match-1",
		GroupID:     "group-1",
		Title:       "Friday football",
		Venue:       "Stadium A",
		ScheduledAt: pgtype.Timestamptz{Time: scheduledAt, Valid: true},
		Status:      "draft",
		CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
	}

	match, err := MatchToDomain(row)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.ID() != "match-1" {
		t.Errorf("expected ID 'match-1', got %q", match.ID())
	}
	if match.Title() != "Friday football" {
		t.Errorf("expected title 'Friday football', got %q", match.Title())
	}
	if match.Status() != entities.MatchStatusDraft {
		t.Errorf("expected status draft, got %q", match.Status())
	}
}

func TestMatchToDomain_ReturnsError_WhenStatusIsInvalid(t *testing.T) {
	t.Parallel()

	row := generated.Matches{
		ID:          "match-1",
		GroupID:     "group-1",
		Title:       "Test",
		Venue:       "Venue",
		ScheduledAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Status:      "not_a_real_status",
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	match, err := MatchToDomain(row)

	if match != nil {
		t.Errorf("expected nil match, got %+v", match)
	}
	if !errors.Is(err, domainerrors.ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestMatchToCreateParams_BuildsParamsFromEntity(t *testing.T) {
	t.Parallel()

	scheduledAt := time.Date(2026, 5, 1, 18, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 11, 11, 0, 0, 0, time.UTC)
	match, err := entities.NewMatch(
		"match-1", "group-1", "Title", "Venue",
		scheduledAt, createdAt,
	)
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}

	params := MatchToCreateParams(match)

	if params.ID != "match-1" {
		t.Errorf("expected ID 'match-1', got %q", params.ID)
	}
	if params.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", params.Status)
	}
	if !params.ScheduledAt.Time.Equal(scheduledAt) {
		t.Errorf("expected scheduledAt %v, got %v", scheduledAt, params.ScheduledAt.Time)
	}
}

func TestMatchToUpdateStatusParams_BuildsParamsFromEntity(t *testing.T) {
	t.Parallel()

	match, _ := entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          "match-1",
		GroupID:     "group-1",
		Title:       "Title",
		Venue:       "Venue",
		ScheduledAt: time.Now(),
		Status:      entities.MatchStatusOpen,
		CreatedAt:   time.Now(),
	})

	params := MatchToUpdateStatusParams(match)

	if params.ID != "match-1" {
		t.Errorf("expected ID 'match-1', got %q", params.ID)
	}
	if params.Status != "open" {
		t.Errorf("expected status 'open', got %q", params.Status)
	}
}
