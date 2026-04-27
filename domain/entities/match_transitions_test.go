package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// newMatchInStatus is a test helper that builds a Match in the given status.
// It panics on construction error because tests should never produce invalid
// inputs; if they do, the test setup itself is broken and we want a fast
// loud failure.
func newMatchInStatus(t *testing.T, status MatchStatus) *Match {
	t.Helper()

	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	scheduled := now.Add(24 * time.Hour)

	match, err := RehydrateMatch(MatchSnapshot{
		ID:          "m-1",
		GroupID:     "g-1",
		Title:       "Title",
		Venue:       "Venue",
		ScheduledAt: scheduled,
		Status:      status,
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("test helper failed to build match in status %q: %v", status, err)
	}
	return match
}

func TestMatch_HappyPathFullLifecycle(t *testing.T) {
	t.Parallel()

	match := newMatchInStatus(t, MatchStatusDraft)

	if err := match.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if match.Status() != MatchStatusOpen {
		t.Errorf("expected status open, got %q", match.Status())
	}

	if err := match.MarkTeamsReady(); err != nil {
		t.Fatalf("MarkTeamsReady failed: %v", err)
	}
	if match.Status() != MatchStatusTeamsReady {
		t.Errorf("expected status teams_ready, got %q", match.Status())
	}

	if err := match.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if match.Status() != MatchStatusInProgress {
		t.Errorf("expected status in_progress, got %q", match.Status())
	}

	if err := match.Complete(); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if match.Status() != MatchStatusCompleted {
		t.Errorf("expected status completed, got %q", match.Status())
	}

	mvp := PlayerID("p-mvp")
	if err := match.Finalize(&mvp); err != nil {
		t.Fatalf("Finalize failed: %v", err)
	}
	if match.Status() != MatchStatusClosed {
		t.Errorf("expected status closed, got %q", match.Status())
	}
	if match.MVP() == nil || *match.MVP() != mvp {
		t.Errorf("expected MVP %q, got %v", mvp, match.MVP())
	}
}

func TestMatch_Finalize_AcceptsNilMVP_WhenNoVotesCast(t *testing.T) {
	t.Parallel()

	match := newMatchInStatus(t, MatchStatusCompleted)

	if err := match.Finalize(nil); err != nil {
		t.Fatalf("Finalize with nil MVP failed: %v", err)
	}
	if match.Status() != MatchStatusClosed {
		t.Errorf("expected status closed, got %q", match.Status())
	}
	if match.MVP() != nil {
		t.Errorf("expected nil MVP, got %v", match.MVP())
	}
}

func TestMatch_TransitionsRejectInvalidSourceStatus(t *testing.T) {
	t.Parallel()

	// Each row defines a transition method and the only status from which
	// it is allowed. Every other status must be rejected.
	transitions := []struct {
		name        string
		allowedFrom MatchStatus
		invoke      func(*Match) error
	}{
		{name: "Open", allowedFrom: MatchStatusDraft, invoke: (*Match).Open},
		{name: "MarkTeamsReady", allowedFrom: MatchStatusOpen, invoke: (*Match).MarkTeamsReady},
		{name: "Start", allowedFrom: MatchStatusTeamsReady, invoke: (*Match).Start},
		{name: "Complete", allowedFrom: MatchStatusInProgress, invoke: (*Match).Complete},
		{name: "Finalize", allowedFrom: MatchStatusCompleted, invoke: func(m *Match) error { return m.Finalize(nil) }},
	}

	allStatuses := []MatchStatus{
		MatchStatusDraft,
		MatchStatusOpen,
		MatchStatusTeamsReady,
		MatchStatusInProgress,
		MatchStatusCompleted,
		MatchStatusClosed,
	}

	for _, transition := range transitions {
		for _, status := range allStatuses {
			if status == transition.allowedFrom {
				continue
			}

			testName := transition.name + "_from_" + string(status)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()

				match := newMatchInStatus(t, status)

				err := transition.invoke(match)

				if err == nil {
					t.Errorf("expected error when calling %s from status %q, got nil", transition.name, status)
				}
				if !errors.Is(err, domainerrors.ErrInvalidTransition) {
					t.Errorf("expected ErrInvalidTransition, got %v", err)
				}
				if match.Status() != status {
					t.Errorf("expected status to remain %q after failed transition, got %q", status, match.Status())
				}
			})
		}
	}
}
