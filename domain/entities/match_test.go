package entities

import (
	"errors"
	"strings"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewMatch_ReturnsMatch_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	scheduledAt := time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	match, err := NewMatch(
		"match-1",
		"group-1",
		"Foot du jeudi soir",
		"Stade de Gerland, Lyon",
		scheduledAt,
		createdAt,
	)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.ID() != "match-1" {
		t.Errorf("expected ID 'match-1', got %q", match.ID())
	}
	if match.GroupID() != "group-1" {
		t.Errorf("expected GroupID 'group-1', got %q", match.GroupID())
	}
	if match.Title() != "Foot du jeudi soir" {
		t.Errorf("expected title 'Foot du jeudi soir', got %q", match.Title())
	}
	if match.Venue() != "Stade de Gerland, Lyon" {
		t.Errorf("expected venue, got %q", match.Venue())
	}
	if !match.ScheduledAt().Equal(scheduledAt) {
		t.Errorf("expected scheduledAt %v, got %v", scheduledAt, match.ScheduledAt())
	}
	if match.Status() != MatchStatusDraft {
		t.Errorf("expected status draft, got %q", match.Status())
	}
	if !match.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, match.CreatedAt())
	}
}

func TestNewMatch_TrimsTitleAndVenue(t *testing.T) {
	t.Parallel()

	match, err := RehydrateMatch(MatchSnapshot{
		ID:          "m-1",
		GroupID:     "g-1",
		Title:       "  Tournoi  ",
		Venue:       "  Lyon  ",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      MatchStatusOpen,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.Title() != "Tournoi" {
		t.Errorf("expected trimmed title 'Tournoi', got %q", match.Title())
	}
	if match.Venue() != "Lyon" {
		t.Errorf("expected trimmed venue 'Lyon', got %q", match.Venue())
	}
}

func TestNewMatch_RejectsInvalidStatusString(t *testing.T) {
	t.Parallel()

	match, err := RehydrateMatch(MatchSnapshot{
		ID:          "m-1",
		GroupID:     "g-1",
		Title:       "Title",
		Venue:       "Venue",
		ScheduledAt: time.Now().Add(time.Hour),
		Status:      MatchStatus("not-a-real-status"),
		CreatedAt:   time.Now(),
	})

	if match != nil {
		t.Errorf("expected nil match, got %+v", match)
	}
	if !errors.Is(err, domainerrors.ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestNewMatch_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	validTime := time.Now()
	longTitle := strings.Repeat("a", 101)
	longVenue := strings.Repeat("a", 201)

	cases := []struct {
		name        string
		id          MatchID
		groupID     GroupID
		title       string
		venue       string
		scheduledAt time.Time
		status      MatchStatus
		createdAt   time.Time
		wantErr     error
	}{
		{name: "empty id", id: "", groupID: "g-1", title: "T", venue: "V", scheduledAt: validTime, status: MatchStatusDraft, createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "empty group id", id: "m-1", groupID: "", title: "T", venue: "V", scheduledAt: validTime, status: MatchStatusDraft, createdAt: validTime, wantErr: domainerrors.ErrInvalidID},
		{name: "empty title", id: "m-1", groupID: "g-1", title: "", venue: "V", scheduledAt: validTime, status: MatchStatusDraft, createdAt: validTime, wantErr: domainerrors.ErrInvalidName},
		{name: "title too long", id: "m-1", groupID: "g-1", title: longTitle, venue: "V", scheduledAt: validTime, status: MatchStatusDraft, createdAt: validTime, wantErr: domainerrors.ErrInvalidName},
		{name: "empty venue", id: "m-1", groupID: "g-1", title: "T", venue: "", scheduledAt: validTime, status: MatchStatusDraft, createdAt: validTime, wantErr: domainerrors.ErrInvalidName},
		{name: "venue too long", id: "m-1", groupID: "g-1", title: "T", venue: longVenue, scheduledAt: validTime, status: MatchStatusDraft, createdAt: validTime, wantErr: domainerrors.ErrInvalidName},
		{name: "zero scheduledAt", id: "m-1", groupID: "g-1", title: "T", venue: "V", scheduledAt: time.Time{}, status: MatchStatusDraft, createdAt: validTime, wantErr: domainerrors.ErrInvalidDate},
		{name: "zero createdAt", id: "m-1", groupID: "g-1", title: "T", venue: "V", scheduledAt: validTime, status: MatchStatusDraft, createdAt: time.Time{}, wantErr: domainerrors.ErrInvalidDate},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			match, err := RehydrateMatch(MatchSnapshot{
				ID:          testCase.id,
				GroupID:     testCase.groupID,
				Title:       testCase.title,
				Venue:       testCase.venue,
				ScheduledAt: testCase.scheduledAt,
				Status:      testCase.status,
				CreatedAt:   testCase.createdAt,
			})

			if match != nil {
				t.Errorf("expected nil match, got %+v", match)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}
