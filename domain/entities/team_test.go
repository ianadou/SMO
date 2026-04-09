package entities

import (
	"errors"
	"strings"
	"testing"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewTeam_ReturnsTeam_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	players := []PlayerID{"p-1", "p-2", "p-3"}

	team, err := NewTeam("team-a", "match-1", "Team A", players)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if team.ID() != "team-a" {
		t.Errorf("expected ID 'team-a', got %q", team.ID())
	}
	if team.MatchID() != "match-1" {
		t.Errorf("expected MatchID 'match-1', got %q", team.MatchID())
	}
	if team.Label() != "Team A" {
		t.Errorf("expected label 'Team A', got %q", team.Label())
	}
	if team.Size() != 3 {
		t.Errorf("expected size 3, got %d", team.Size())
	}
}

func TestNewTeam_AcceptsEmptyPlayerList(t *testing.T) {
	t.Parallel()

	team, err := NewTeam("team-a", "match-1", "Team A", []PlayerID{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if team.Size() != 0 {
		t.Errorf("expected size 0, got %d", team.Size())
	}
}

func TestTeam_PlayersReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	originalPlayers := []PlayerID{"p-1", "p-2"}
	team, _ := NewTeam("t-1", "m-1", "Team A", originalPlayers)

	// arrange: caller mutates the original slice they passed in
	originalPlayers[0] = "HACKED"

	// act: read the team's players
	got := team.Players()

	// assert: team's internal state is not affected
	if got[0] != "p-1" {
		t.Errorf("expected team to keep 'p-1', got %q (defensive copy failed)", got[0])
	}

	// arrange: caller mutates the slice returned by Players()
	got[0] = "HACKED-AGAIN"

	// act: read again
	gotAgain := team.Players()

	// assert: team's internal state is still safe
	if gotAgain[0] != "p-1" {
		t.Errorf("expected team to keep 'p-1', got %q (returned slice not isolated)", gotAgain[0])
	}
}

func TestNewTeam_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	longLabel := strings.Repeat("a", 21)

	cases := []struct {
		name    string
		id      TeamID
		matchID MatchID
		label   string
		wantErr error
	}{
		{name: "empty id", id: "", matchID: "m-1", label: "Team A", wantErr: domainerrors.ErrInvalidID},
		{name: "empty match id", id: "t-1", matchID: "", label: "Team A", wantErr: domainerrors.ErrInvalidID},
		{name: "empty label", id: "t-1", matchID: "m-1", label: "", wantErr: domainerrors.ErrInvalidName},
		{name: "whitespace-only label", id: "t-1", matchID: "m-1", label: "   ", wantErr: domainerrors.ErrInvalidName},
		{name: "label too long", id: "t-1", matchID: "m-1", label: longLabel, wantErr: domainerrors.ErrInvalidName},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			team, err := NewTeam(testCase.id, testCase.matchID, testCase.label, []PlayerID{})

			if team != nil {
				t.Errorf("expected nil team, got %+v", team)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}
