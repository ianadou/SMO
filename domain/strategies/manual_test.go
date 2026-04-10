package strategies

import (
	"errors"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// newTestPlayer is a test helper that builds a Player with a fixed group
// and ranking, panicking on construction error. Used to keep test setup
// concise when the player details are not the focus of the test.
func newTestPlayer(t *testing.T, id entities.PlayerID) *entities.Player {
	t.Helper()

	player, err := entities.NewPlayer(id, "test-group", "Test "+string(id), 1000)
	if err != nil {
		t.Fatalf("test helper failed to build player %q: %v", id, err)
	}
	return player
}

func TestNewManualAssignmentStrategy_ReturnsStrategy_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	teamA := []entities.PlayerID{"p-1", "p-2", "p-3"}
	teamB := []entities.PlayerID{"p-4", "p-5", "p-6"}

	strategy, err := NewManualAssignmentStrategy(teamA, teamB)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if strategy == nil {
		t.Fatal("expected non-nil strategy")
	}
}

func TestNewManualAssignmentStrategy_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		teamA []entities.PlayerID
		teamB []entities.PlayerID
	}{
		{name: "empty team A", teamA: []entities.PlayerID{}, teamB: []entities.PlayerID{"p-1"}},
		{name: "empty team B", teamA: []entities.PlayerID{"p-1"}, teamB: []entities.PlayerID{}},
		{name: "both teams empty", teamA: []entities.PlayerID{}, teamB: []entities.PlayerID{}},
		{name: "duplicate player across teams", teamA: []entities.PlayerID{"p-1", "p-2"}, teamB: []entities.PlayerID{"p-2", "p-3"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			strategy, err := NewManualAssignmentStrategy(testCase.teamA, testCase.teamB)

			if strategy != nil {
				t.Errorf("expected nil strategy, got %+v", strategy)
			}
			if !errors.Is(err, domainerrors.ErrInvalidAssignment) {
				t.Errorf("expected ErrInvalidAssignment, got %v", err)
			}
		})
	}
}

func TestManualAssignmentStrategy_Assign_ReturnsPredefinedTeams_WhenInputMatches(t *testing.T) {
	t.Parallel()

	players := []*entities.Player{
		newTestPlayer(t, "p-1"),
		newTestPlayer(t, "p-2"),
		newTestPlayer(t, "p-3"),
		newTestPlayer(t, "p-4"),
	}

	strategy, _ := NewManualAssignmentStrategy(
		[]entities.PlayerID{"p-1", "p-3"},
		[]entities.PlayerID{"p-2", "p-4"},
	)

	teamA, teamB, err := strategy.Assign(players)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(teamA) != 2 || teamA[0] != "p-1" || teamA[1] != "p-3" {
		t.Errorf("expected team A [p-1 p-3], got %v", teamA)
	}
	if len(teamB) != 2 || teamB[0] != "p-2" || teamB[1] != "p-4" {
		t.Errorf("expected team B [p-2 p-4], got %v", teamB)
	}
}

func TestManualAssignmentStrategy_Assign_ReturnsError_WhenInputDoesNotMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		players []*entities.Player
		teamA   []entities.PlayerID
		teamB   []entities.PlayerID
	}{
		{
			name:    "input has fewer players than assignment",
			players: []*entities.Player{newTestPlayer(t, "p-1"), newTestPlayer(t, "p-2")},
			teamA:   []entities.PlayerID{"p-1"},
			teamB:   []entities.PlayerID{"p-2", "p-3"},
		},
		{
			name:    "input has more players than assignment",
			players: []*entities.Player{newTestPlayer(t, "p-1"), newTestPlayer(t, "p-2"), newTestPlayer(t, "p-3")},
			teamA:   []entities.PlayerID{"p-1"},
			teamB:   []entities.PlayerID{"p-2"},
		},
		{
			name:    "input has different player ids than assignment",
			players: []*entities.Player{newTestPlayer(t, "p-1"), newTestPlayer(t, "p-2")},
			teamA:   []entities.PlayerID{"p-3"},
			teamB:   []entities.PlayerID{"p-4"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			strategy, _ := NewManualAssignmentStrategy(testCase.teamA, testCase.teamB)

			teamA, teamB, err := strategy.Assign(testCase.players)

			if teamA != nil || teamB != nil {
				t.Errorf("expected nil teams on error, got A=%v B=%v", teamA, teamB)
			}
			if !errors.Is(err, domainerrors.ErrInvalidAssignment) {
				t.Errorf("expected ErrInvalidAssignment, got %v", err)
			}
		})
	}
}

func TestManualAssignmentStrategy_Assign_DefensiveCopyOnReturn(t *testing.T) {
	t.Parallel()

	players := []*entities.Player{
		newTestPlayer(t, "p-1"),
		newTestPlayer(t, "p-2"),
	}

	strategy, _ := NewManualAssignmentStrategy(
		[]entities.PlayerID{"p-1"},
		[]entities.PlayerID{"p-2"},
	)

	teamA, _, _ := strategy.Assign(players)

	// Mutate the returned slice and verify the strategy state is unchanged.
	teamA[0] = "HACKED"

	teamAAgain, _, _ := strategy.Assign(players)
	if teamAAgain[0] != "p-1" {
		t.Errorf("expected strategy state to remain 'p-1', got %q", teamAAgain[0])
	}
}
