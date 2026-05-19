package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestMatch_Complete_RecordsScoreAndTransitions_WhenInProgress(t *testing.T) {
	t.Parallel()

	match := newMatchInStatus(t, MatchStatusInProgress)

	err := match.Complete(3, 1)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if match.Status() != MatchStatusCompleted {
		t.Errorf("expected status completed, got %q", match.Status())
	}
	if match.ScoreA() == nil || *match.ScoreA() != 3 {
		t.Errorf("expected score A 3, got %v", match.ScoreA())
	}
	if match.ScoreB() == nil || *match.ScoreB() != 1 {
		t.Errorf("expected score B 1, got %v", match.ScoreB())
	}
}

func TestMatch_Complete_RejectsNegativeScore_WithoutTransitioning(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		scoreA, scoreB int
	}{
		{name: "negative A", scoreA: -1, scoreB: 0},
		{name: "negative B", scoreA: 0, scoreB: -2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			match := newMatchInStatus(t, MatchStatusInProgress)

			err := match.Complete(tc.scoreA, tc.scoreB)

			if !errors.Is(err, domainerrors.ErrInvalidMatchScore) {
				t.Fatalf("expected ErrInvalidMatchScore, got %v", err)
			}
			if match.Status() != MatchStatusInProgress {
				t.Errorf("expected status to remain in_progress, got %q", match.Status())
			}
			if match.ScoreA() != nil || match.ScoreB() != nil {
				t.Errorf("expected no score recorded, got %v / %v", match.ScoreA(), match.ScoreB())
			}
		})
	}
}

func TestMatch_WinningSide_DerivesFromScore(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		scoreA, scoreB int
		want           *TeamSide
	}{
		{name: "team A wins", scoreA: 4, scoreB: 2, want: ptrSide(TeamSideA)},
		{name: "team B wins", scoreA: 1, scoreB: 5, want: ptrSide(TeamSideB)},
		{name: "draw is no winner", scoreA: 3, scoreB: 3, want: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			match := newMatchInStatus(t, MatchStatusInProgress)
			if err := match.Complete(tc.scoreA, tc.scoreB); err != nil {
				t.Fatalf("Complete failed: %v", err)
			}

			got := match.WinningSide()

			switch {
			case tc.want == nil && got != nil:
				t.Errorf("expected no winner, got %q", *got)
			case tc.want != nil && got == nil:
				t.Errorf("expected winner %q, got nil", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Errorf("expected winner %q, got %q", *tc.want, *got)
			}
		})
	}
}

func TestMatch_WinningSide_ReturnsNil_WhenNoScoreRecorded(t *testing.T) {
	t.Parallel()

	match := newMatchInStatus(t, MatchStatusInProgress)

	if got := match.WinningSide(); got != nil {
		t.Errorf("expected nil winner before completion, got %q", *got)
	}
}

func TestRehydrateMatch_RoundTripsScore(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	a, b := 2, 0

	match, err := RehydrateMatch(MatchSnapshot{
		ID:          "m-1",
		GroupID:     "g-1",
		Title:       "Title",
		Venue:       "Venue",
		ScheduledAt: now.Add(24 * time.Hour),
		Status:      MatchStatusCompleted,
		ScoreA:      &a,
		ScoreB:      &b,
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("RehydrateMatch failed: %v", err)
	}
	if match.ScoreA() == nil || *match.ScoreA() != 2 {
		t.Errorf("expected score A 2, got %v", match.ScoreA())
	}
	if match.ScoreB() == nil || *match.ScoreB() != 0 {
		t.Errorf("expected score B 0, got %v", match.ScoreB())
	}
	if side := match.WinningSide(); side == nil || *side != TeamSideA {
		t.Errorf("expected winning side A, got %v", side)
	}
}

func ptrSide(s TeamSide) *TeamSide { return &s }
