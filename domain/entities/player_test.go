package entities

import (
	"errors"
	"strings"
	"testing"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewPlayer_ReturnsPlayer_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	player, err := NewPlayer("player-1", "group-1", "Eddin", 1200)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if player.ID() != "player-1" {
		t.Errorf("expected ID 'player-1', got %q", player.ID())
	}
	if player.GroupID() != "group-1" {
		t.Errorf("expected GroupID 'group-1', got %q", player.GroupID())
	}
	if player.Name() != "Eddin" {
		t.Errorf("expected name 'Eddin', got %q", player.Name())
	}
	if player.Ranking() != 1200 {
		t.Errorf("expected ranking 1200, got %d", player.Ranking())
	}
}

func TestNewPlayer_TrimsWhitespaceAroundName(t *testing.T) {
	t.Parallel()

	player, err := NewPlayer("p-1", "g-1", "  Eddin  ", DefaultPlayerRanking())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if player.Name() != "Eddin" {
		t.Errorf("expected trimmed name 'Eddin', got %q", player.Name())
	}
}

func TestNewPlayer_AcceptsAnyRankingValue(t *testing.T) {
	t.Parallel()

	cases := []int{0, -100, 1000, 9999}

	for _, ranking := range cases {
		_, err := NewPlayer("p-1", "g-1", "Test", ranking)
		if err != nil {
			t.Errorf("expected no error for ranking %d, got: %v", ranking, err)
		}
	}
}

func TestDefaultPlayerRanking_ReturnsExpectedValue(t *testing.T) {
	t.Parallel()

	if DefaultPlayerRanking() != 1000 {
		t.Errorf("expected default ranking 1000, got %d", DefaultPlayerRanking())
	}
}

func TestNewPlayer_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	longName := strings.Repeat("a", 51)

	cases := []struct {
		name       string
		id         PlayerID
		groupID    GroupID
		playerName string
		wantErr    error
	}{
		{name: "empty id", id: "", groupID: "g-1", playerName: "Test", wantErr: domainerrors.ErrInvalidID},
		{name: "empty group id", id: "p-1", groupID: "", playerName: "Test", wantErr: domainerrors.ErrInvalidID},
		{name: "empty name", id: "p-1", groupID: "g-1", playerName: "", wantErr: domainerrors.ErrInvalidName},
		{name: "whitespace-only name", id: "p-1", groupID: "g-1", playerName: "   ", wantErr: domainerrors.ErrInvalidName},
		{name: "name too long", id: "p-1", groupID: "g-1", playerName: longName, wantErr: domainerrors.ErrInvalidName},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			player, err := NewPlayer(testCase.id, testCase.groupID, testCase.playerName, 1000)

			if player != nil {
				t.Errorf("expected nil player, got %+v", player)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}
