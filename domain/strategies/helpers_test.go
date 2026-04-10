package strategies

import (
	"fmt"
	"testing"

	"github.com/ianadou/smo/domain/entities"
)

// buildTestPlayers builds a slice of n test players with sequential IDs
// (test-player-0, test-player-1, ...) and a default ranking. Useful when
// the test does not care about the exact player data.
func buildTestPlayers(t *testing.T, count int) []*entities.Player {
	t.Helper()

	players := make([]*entities.Player, 0, count)
	for index := 0; index < count; index++ {
		id := entities.PlayerID(fmt.Sprintf("test-player-%d", index))
		player, err := entities.NewPlayer(id, "test-group", fmt.Sprintf("Player %d", index), 1000)
		if err != nil {
			t.Fatalf("buildTestPlayers failed at index %d: %v", index, err)
		}
		players = append(players, player)
	}
	return players
}

// newTestPlayerWithRanking is a test helper that builds a Player with a
// specific ranking, useful when the ranking value matters for the test.
func newTestPlayerWithRanking(t *testing.T, id entities.PlayerID, ranking int) *entities.Player {
	t.Helper()

	player, err := entities.NewPlayer(id, "test-group", "Test "+string(id), ranking)
	if err != nil {
		t.Fatalf("newTestPlayerWithRanking failed for %q: %v", id, err)
	}
	return player
}

// slicesEqual reports whether two PlayerID slices have the same length
// and the same elements in the same order.
func slicesEqual(a, b []entities.PlayerID) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}

// totalRankingForIDs sums the rankings of the players whose IDs are in
// the given list. Used to verify that ranking-based strategies produce
// balanced teams.
func totalRankingForIDs(t *testing.T, players []*entities.Player, ids []entities.PlayerID) int {
	t.Helper()

	rankingByID := make(map[entities.PlayerID]int, len(players))
	for _, player := range players {
		rankingByID[player.ID()] = player.Ranking()
	}

	total := 0
	for _, id := range ids {
		ranking, ok := rankingByID[id]
		if !ok {
			t.Fatalf("player %q not found in players list", id)
		}
		total += ranking
	}
	return total
}
