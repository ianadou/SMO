package strategies

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// newDeterministicRng builds a *rand.Rand with a fixed seed so that
// random-based tests are reproducible across runs and machines.
func newDeterministicRng() *rand.Rand {
	return rand.New(rand.NewPCG(42, 1024))
}

func TestNewRandomAssignmentStrategy_ReturnsStrategy_WhenRngIsProvided(t *testing.T) {
	t.Parallel()

	strategy, err := NewRandomAssignmentStrategy(newDeterministicRng())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if strategy == nil {
		t.Fatal("expected non-nil strategy")
	}
}

func TestNewRandomAssignmentStrategy_ReturnsError_WhenRngIsNil(t *testing.T) {
	t.Parallel()

	strategy, err := NewRandomAssignmentStrategy(nil)

	if strategy != nil {
		t.Errorf("expected nil strategy, got %+v", strategy)
	}
	if !errors.Is(err, domainerrors.ErrInvalidAssignment) {
		t.Errorf("expected ErrInvalidAssignment, got %v", err)
	}
}

func TestRandomAssignmentStrategy_Assign_DistributesAllPlayersExactlyOnce(t *testing.T) {
	t.Parallel()

	players := buildTestPlayers(t, 10)
	strategy, _ := NewRandomAssignmentStrategy(newDeterministicRng())

	teamA, teamB, err := strategy.Assign(players)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Property: every player must appear exactly once across both teams.
	seen := make(map[entities.PlayerID]int)
	for _, id := range teamA {
		seen[id]++
	}
	for _, id := range teamB {
		seen[id]++
	}

	if len(seen) != len(players) {
		t.Errorf("expected %d distinct players in assignment, got %d", len(players), len(seen))
	}
	for id, count := range seen {
		if count != 1 {
			t.Errorf("player %q appears %d times, expected exactly 1", id, count)
		}
	}
}

func TestRandomAssignmentStrategy_Assign_SplitsEvenCountInHalf(t *testing.T) {
	t.Parallel()

	players := buildTestPlayers(t, 10)
	strategy, _ := NewRandomAssignmentStrategy(newDeterministicRng())

	teamA, teamB, _ := strategy.Assign(players)

	if len(teamA) != 5 {
		t.Errorf("expected team A size 5, got %d", len(teamA))
	}
	if len(teamB) != 5 {
		t.Errorf("expected team B size 5, got %d", len(teamB))
	}
}

func TestRandomAssignmentStrategy_Assign_GivesExtraPlayerToTeamAOnOddCount(t *testing.T) {
	t.Parallel()

	players := buildTestPlayers(t, 7)
	strategy, _ := NewRandomAssignmentStrategy(newDeterministicRng())

	teamA, teamB, _ := strategy.Assign(players)

	if len(teamA) != 4 {
		t.Errorf("expected team A size 4 (extra player), got %d", len(teamA))
	}
	if len(teamB) != 3 {
		t.Errorf("expected team B size 3, got %d", len(teamB))
	}
}

func TestRandomAssignmentStrategy_Assign_IsReproducibleWithSameSeed(t *testing.T) {
	t.Parallel()

	players := buildTestPlayers(t, 10)

	strategy1, _ := NewRandomAssignmentStrategy(newDeterministicRng())
	strategy2, _ := NewRandomAssignmentStrategy(newDeterministicRng())

	teamA1, teamB1, _ := strategy1.Assign(players)
	teamA2, teamB2, _ := strategy2.Assign(players)

	if !slicesEqual(teamA1, teamA2) {
		t.Errorf("expected reproducible team A, got %v vs %v", teamA1, teamA2)
	}
	if !slicesEqual(teamB1, teamB2) {
		t.Errorf("expected reproducible team B, got %v vs %v", teamB1, teamB2)
	}
}

func TestRandomAssignmentStrategy_Assign_DoesNotMutateInputSlice(t *testing.T) {
	t.Parallel()

	players := buildTestPlayers(t, 6)
	originalIDs := make([]entities.PlayerID, len(players))
	for index, player := range players {
		originalIDs[index] = player.ID()
	}

	strategy, _ := NewRandomAssignmentStrategy(newDeterministicRng())
	_, _, _ = strategy.Assign(players)

	for index, player := range players {
		if player.ID() != originalIDs[index] {
			t.Errorf("input slice was mutated at index %d: expected %q, got %q", index, originalIDs[index], player.ID())
		}
	}
}

func TestRandomAssignmentStrategy_Assign_ReturnsError_WhenTooFewPlayers(t *testing.T) {
	t.Parallel()

	cases := []int{0, 1}

	for _, count := range cases {
		players := buildTestPlayers(t, count)
		strategy, _ := NewRandomAssignmentStrategy(newDeterministicRng())

		teamA, teamB, err := strategy.Assign(players)

		if teamA != nil || teamB != nil {
			t.Errorf("expected nil teams on error for %d players", count)
		}
		if !errors.Is(err, domainerrors.ErrInvalidAssignment) {
			t.Errorf("expected ErrInvalidAssignment for %d players, got %v", count, err)
		}
	}
}
