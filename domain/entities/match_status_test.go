package entities

import (
	"errors"
	"testing"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestParseMatchStatus_ReturnsTypedStatus_WhenValueIsKnown(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want MatchStatus
	}{
		{name: "draft", raw: "draft", want: MatchStatusDraft},
		{name: "open", raw: "open", want: MatchStatusOpen},
		{name: "teams_ready", raw: "teams_ready", want: MatchStatusTeamsReady},
		{name: "in_progress", raw: "in_progress", want: MatchStatusInProgress},
		{name: "completed", raw: "completed", want: MatchStatusCompleted},
		{name: "closed", raw: "closed", want: MatchStatusClosed},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMatchStatus(testCase.raw)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got != testCase.want {
				t.Errorf("expected %q, got %q", testCase.want, got)
			}
		})
	}
}

func TestParseMatchStatus_ReturnsError_WhenValueIsUnknown(t *testing.T) {
	t.Parallel()

	cases := []string{
		"",
		"DRAFT",
		"unknown",
		"in-progress",
		"teamsready",
	}

	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMatchStatus(raw)

			if got != "" {
				t.Errorf("expected empty status, got %q", got)
			}
			if !errors.Is(err, domainerrors.ErrInvalidStatus) {
				t.Errorf("expected ErrInvalidStatus, got %v", err)
			}
		})
	}
}
