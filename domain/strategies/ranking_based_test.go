package strategies

import (
	"errors"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewRankingBasedStrategy_ReturnsStrategy(t *testing.T) {
	t.Parallel()

	strategy := NewRankingBasedStrategy()

	if strategy == nil {
		t.Fatal("expected non-nil strategy")
	}
}

func TestRankingBasedStrategy_Assign_DistributesUsingSnakeDraft(t *testing.T) {
	t.Parallel()

	// Build 8 players with rankings 800..100, descending. After sorting,
	// the top player is p-800, then p-700, ..., bottom is p-100.
	//
	// Expected snake draft:
	//   pair 0 (A starts): A=p-800, B=p-700
	//   pair 1 (B starts): B=p-600, A=p-500
	//   pair 2 (A starts): A=p-400, B=p-300
	//   pair 3 (B starts): B=p-200, A=p-100
	//
	// Final teams:
	//   A = [p-800, p-500, p-400, p-100]
	//   B = [p-700, p-600, p-300, p-200]
	players := []*entities.Player{
		newTestPlayerWithRanking(t, "p-100", 100),
		newTestPlayerWithRanking(t, "p-200", 200),
		newTestPlayerWithRanking(t, "p-300", 300),
		newTestPlayerWithRanking(t, "p-400", 400),
		newTestPlayerWithRanking(t, "p-500", 500),
		newTestPlayerWithRanking(t, "p-600", 600),
		newTestPlayerWithRanking(t, "p-700", 700),
		newTestPlayerWithRanking(t, "p-800", 800),
	}

	strategy := NewRankingBasedStrategy()

	teamA, teamB, err := strategy.Assign(players)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedA := []entities.PlayerID{"p-800", "p-500", "p-400", "p-100"}
	expectedB := []entities.PlayerID{"p-700", "p-600", "p-300", "p-200"}

	if !slicesEqual(teamA, expectedA) {
		t.Errorf("expected team A %v, got %v", expectedA, teamA)
	}
	if !slicesEqual(teamB, expectedB) {
		t.Errorf("expected team B %v, got %v", expectedB, teamB)
	}
}

func TestRankingBasedStrategy_Assign_BalancesTotalRanking(t *testing.T) {
	t.Parallel()

	// With 6 players ranked 100..600, the snake draft should produce
	// near-equal total rankings between teams.
	players := []*entities.Player{
		newTestPlayerWithRanking(t, "p-1", 600),
		newTestPlayerWithRanking(t, "p-2", 500),
		newTestPlayerWithRanking(t, "p-3", 400),
		newTestPlayerWithRanking(t, "p-4", 300),
		newTestPlayerWithRanking(t, "p-5", 200),
		newTestPlayerWithRanking(t, "p-6", 100),
	}

	strategy := NewRankingBasedStrategy()

	teamA, teamB, err := strategy.Assign(players)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	totalA := totalRankingForIDs(t, players, teamA)
	totalB := totalRankingForIDs(t, players, teamB)

	// Snake draft on (600, 500, 400, 300, 200, 100):
	//   A = 600 + 300 + 200 = 1100
	//   B = 500 + 400 + 100 = 1000
	// Difference of 100 out of 2100 total = ~4.7%, which is acceptable.
	difference := totalA - totalB
	if difference < 0 {
		difference = -difference
	}
	if difference > 200 {
		t.Errorf("expected balanced totals, got A=%d B=%d (diff %d)", totalA, totalB, difference)
	}
}

func TestRankingBasedStrategy_Assign_SplitsEvenCountInHalf(t *testing.T) {
	t.Parallel()

	players := buildTestPlayers(t, 10)
	strategy := NewRankingBasedStrategy()

	teamA, teamB, _ := strategy.Assign(players)

	if len(teamA) != 5 {
		t.Errorf("expected team A size 5, got %d", len(teamA))
	}
	if len(teamB) != 5 {
		t.Errorf("expected team B size 5, got %d", len(teamB))
	}
}

func TestRankingBasedStrategy_Assign_GivesExtraPlayerToTeamAOnOddCount(t *testing.T) {
	t.Parallel()

	players := buildTestPlayers(t, 5)
	strategy := NewRankingBasedStrategy()

	teamA, teamB, _ := strategy.Assign(players)

	if len(teamA) != 3 {
		t.Errorf("expected team A size 3 (extra player), got %d", len(teamA))
	}
	if len(teamB) != 2 {
		t.Errorf("expected team B size 2, got %d", len(teamB))
	}
}

func TestRankingBasedStrategy_Assign_DoesNotMutateInputSlice(t *testing.T) {
	t.Parallel()

	players := []*entities.Player{
		newTestPlayerWithRanking(t, "p-1", 100),
		newTestPlayerWithRanking(t, "p-2", 500),
		newTestPlayerWithRanking(t, "p-3", 300),
		newTestPlayerWithRanking(t, "p-4", 200),
	}

	originalIDs := make([]entities.PlayerID, len(players))
	for index, player := range players {
		originalIDs[index] = player.ID()
	}

	strategy := NewRankingBasedStrategy()
	_, _, _ = strategy.Assign(players)

	for index, player := range players {
		if player.ID() != originalIDs[index] {
			t.Errorf("input slice was mutated at index %d: expected %q, got %q", index, originalIDs[index], player.ID())
		}
	}
}

func TestRankingBasedStrategy_Assign_ReturnsError_WhenTooFewPlayers(t *testing.T) {
	t.Parallel()

	cases := []int{0, 1}

	for _, count := range cases {
		players := buildTestPlayers(t, count)
		strategy := NewRankingBasedStrategy()

		teamA, teamB, err := strategy.Assign(players)

		if teamA != nil || teamB != nil {
			t.Errorf("expected nil teams on error for %d players", count)
		}
		if !errors.Is(err, domainerrors.ErrInvalidAssignment) {
			t.Errorf("expected ErrInvalidAssignment for %d players, got %v", count, err)
		}
	}
}
