package ranking

import (
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// newTestPlayer is a test helper that builds a Player with a specific
// ranking, panicking on construction error.
func newTestPlayer(t *testing.T, id entities.PlayerID, ranking int) *entities.Player {
	t.Helper()

	player, err := entities.NewPlayer(id, "test-group", "Test "+string(id), ranking)
	if err != nil {
		t.Fatalf("test helper failed to build player %q: %v", id, err)
	}
	return player
}

// newTestVote is a test helper that builds a Vote, panicking on
// construction error.
func newTestVote(t *testing.T, id entities.VoteID, voterID, votedID entities.PlayerID, score int) *entities.Vote {
	t.Helper()

	vote, err := entities.NewVote(id, "test-match", voterID, votedID, score, time.Now())
	if err != nil {
		t.Fatalf("test helper failed to build vote %q: %v", id, err)
	}
	return vote
}

func TestNewCalculator_ReturnsCalculator_WhenLearningRateIsValid(t *testing.T) {
	t.Parallel()

	cases := []float64{0.01, 0.1, 0.5, 1.0}

	for _, rate := range cases {
		calc, err := NewCalculator(rate)
		if err != nil {
			t.Errorf("expected no error for rate %v, got: %v", rate, err)
		}
		if calc == nil {
			t.Errorf("expected non-nil calculator for rate %v", rate)
		}
	}
}

func TestNewCalculator_ReturnsError_WhenLearningRateIsInvalid(t *testing.T) {
	t.Parallel()

	cases := []float64{-1.0, -0.1, 0, 1.01, 2.0, 100.0}

	for _, rate := range cases {
		calc, err := NewCalculator(rate)
		if calc != nil {
			t.Errorf("expected nil calculator for rate %v, got %+v", rate, calc)
		}
		if !errors.Is(err, domainerrors.ErrInvalidParameter) {
			t.Errorf("expected ErrInvalidParameter for rate %v, got %v", rate, err)
		}
	}
}

func TestDefaultLearningRate_ReturnsExpectedValue(t *testing.T) {
	t.Parallel()

	if DefaultLearningRate() != 0.1 {
		t.Errorf("expected default learning rate 0.1, got %v", DefaultLearningRate())
	}
}

func TestCompute_ReturnsEmptyMap_WhenNoPlayers(t *testing.T) {
	t.Parallel()

	calc, _ := NewCalculator(DefaultLearningRate())

	result, err := calc.Compute([]*entities.Player{}, []*entities.Vote{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result map, got %d entries", len(result))
	}
}

func TestCompute_KeepsRankingUnchanged_WhenPlayerReceivedNoVote(t *testing.T) {
	t.Parallel()

	players := []*entities.Player{
		newTestPlayer(t, "p-1", 1000),
		newTestPlayer(t, "p-2", 1500),
	}
	votes := []*entities.Vote{}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, err := calc.Compute(players, votes)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result["p-1"] != 1000 {
		t.Errorf("expected p-1 ranking 1000, got %d", result["p-1"])
	}
	if result["p-2"] != 1500 {
		t.Errorf("expected p-2 ranking 1500, got %d", result["p-2"])
	}
}

func TestCompute_KeepsRankingUnchanged_WhenAverageScoreIsNeutral(t *testing.T) {
	t.Parallel()

	// p-1 receives a single vote of 3 (neutral). Expected delta is 0.
	players := []*entities.Player{newTestPlayer(t, "p-1", 1000)}
	votes := []*entities.Vote{newTestVote(t, "v-1", "p-2", "p-1", 3)}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, _ := calc.Compute(players, votes)

	if result["p-1"] != 1000 {
		t.Errorf("expected p-1 ranking 1000 (neutral vote), got %d", result["p-1"])
	}
}

func TestCompute_IncreasesRanking_WhenAverageScoreIsAboveNeutral(t *testing.T) {
	t.Parallel()

	// p-1 receives a single vote of 5 (perfect).
	// Centered: 5 - 3 = 2
	// Raw: 2 * 100 = 200
	// Adjusted: 200 * 0.1 = 20
	// New ranking: 1000 + 20 = 1020
	players := []*entities.Player{newTestPlayer(t, "p-1", 1000)}
	votes := []*entities.Vote{newTestVote(t, "v-1", "p-2", "p-1", 5)}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, _ := calc.Compute(players, votes)

	if result["p-1"] != 1020 {
		t.Errorf("expected p-1 ranking 1020 (perfect vote), got %d", result["p-1"])
	}
}

func TestCompute_DecreasesRanking_WhenAverageScoreIsBelowNeutral(t *testing.T) {
	t.Parallel()

	// p-1 receives a single vote of 1 (worst).
	// Centered: 1 - 3 = -2
	// Raw: -2 * 100 = -200
	// Adjusted: -200 * 0.1 = -20
	// New ranking: 1000 - 20 = 980
	players := []*entities.Player{newTestPlayer(t, "p-1", 1000)}
	votes := []*entities.Vote{newTestVote(t, "v-1", "p-2", "p-1", 1)}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, _ := calc.Compute(players, votes)

	if result["p-1"] != 980 {
		t.Errorf("expected p-1 ranking 980 (worst vote), got %d", result["p-1"])
	}
}

func TestCompute_AveragesMultipleVotes(t *testing.T) {
	t.Parallel()

	// p-1 receives three votes: 5, 4, 3.
	// Average: (5+4+3)/3 = 4
	// Centered: 4 - 3 = 1
	// Raw: 1 * 100 = 100
	// Adjusted: 100 * 0.1 = 10
	// New ranking: 1000 + 10 = 1010
	players := []*entities.Player{newTestPlayer(t, "p-1", 1000)}
	votes := []*entities.Vote{
		newTestVote(t, "v-1", "p-2", "p-1", 5),
		newTestVote(t, "v-2", "p-3", "p-1", 4),
		newTestVote(t, "v-3", "p-4", "p-1", 3),
	}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, _ := calc.Compute(players, votes)

	if result["p-1"] != 1010 {
		t.Errorf("expected p-1 ranking 1010 (avg 4), got %d", result["p-1"])
	}
}

func TestCompute_ClampsRankingAtZero(t *testing.T) {
	t.Parallel()

	// p-1 starts at 10 and receives a vote of 1.
	// Adjustment: -20
	// Raw new ranking: 10 - 20 = -10
	// Clamped to 0.
	players := []*entities.Player{newTestPlayer(t, "p-1", 10)}
	votes := []*entities.Vote{newTestVote(t, "v-1", "p-2", "p-1", 1)}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, _ := calc.Compute(players, votes)

	if result["p-1"] != 0 {
		t.Errorf("expected p-1 ranking clamped to 0, got %d", result["p-1"])
	}
}

func TestCompute_AppliesDifferentLearningRates(t *testing.T) {
	t.Parallel()

	// Same input, different learning rates produce different magnitudes.
	players := []*entities.Player{newTestPlayer(t, "p-1", 1000)}
	votes := []*entities.Vote{newTestVote(t, "v-1", "p-2", "p-1", 5)}

	cases := []struct {
		rate         float64
		expectedRank int
	}{
		{rate: 0.1, expectedRank: 1020},  // perfect: 1000 + 200 * 0.1
		{rate: 0.5, expectedRank: 1100},  // perfect: 1000 + 200 * 0.5
		{rate: 1.0, expectedRank: 1200},  // perfect: 1000 + 200 * 1.0
		{rate: 0.05, expectedRank: 1010}, // perfect: 1000 + 200 * 0.05
	}

	for _, testCase := range cases {
		calc, _ := NewCalculator(testCase.rate)
		result, _ := calc.Compute(players, votes)

		if result["p-1"] != testCase.expectedRank {
			t.Errorf("rate %v: expected ranking %d, got %d", testCase.rate, testCase.expectedRank, result["p-1"])
		}
	}
}

func TestCompute_HandlesMultiplePlayers(t *testing.T) {
	t.Parallel()

	// Three players, each with different votes:
	// - p-1: avg 5 → +20
	// - p-2: avg 1 → -20
	// - p-3: no votes → unchanged
	players := []*entities.Player{
		newTestPlayer(t, "p-1", 1000),
		newTestPlayer(t, "p-2", 1500),
		newTestPlayer(t, "p-3", 800),
	}
	votes := []*entities.Vote{
		newTestVote(t, "v-1", "p-2", "p-1", 5),
		newTestVote(t, "v-2", "p-1", "p-2", 1),
	}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, _ := calc.Compute(players, votes)

	if result["p-1"] != 1020 {
		t.Errorf("expected p-1 ranking 1020, got %d", result["p-1"])
	}
	if result["p-2"] != 1480 {
		t.Errorf("expected p-2 ranking 1480, got %d", result["p-2"])
	}
	if result["p-3"] != 800 {
		t.Errorf("expected p-3 ranking 800 (no votes), got %d", result["p-3"])
	}
}

func TestCompute_ReturnsError_WhenVoteTargetsUnknownPlayer(t *testing.T) {
	t.Parallel()

	players := []*entities.Player{newTestPlayer(t, "p-1", 1000)}
	votes := []*entities.Vote{newTestVote(t, "v-1", "p-1", "ghost-player", 5)}

	calc, _ := NewCalculator(DefaultLearningRate())
	result, err := calc.Compute(players, votes)

	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
	if !errors.Is(err, domainerrors.ErrInvalidParameter) {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestCompute_DoesNotMutateInputPlayers(t *testing.T) {
	t.Parallel()

	players := []*entities.Player{newTestPlayer(t, "p-1", 1000)}
	votes := []*entities.Vote{newTestVote(t, "v-1", "p-2", "p-1", 5)}

	calc, _ := NewCalculator(DefaultLearningRate())
	_, _ = calc.Compute(players, votes)

	// Verify the input player still has its original ranking.
	if players[0].Ranking() != 1000 {
		t.Errorf("expected input player ranking unchanged at 1000, got %d", players[0].Ranking())
	}
}
