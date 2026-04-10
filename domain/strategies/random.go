package strategies

import (
	"fmt"
	"math/rand/v2"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

const minPlayersForAssignment = 2

// RandomAssignmentStrategy distributes players randomly into two teams.
//
// The randomness source is injected at construction time so that tests
// can use a seeded source for deterministic results. In production, the
// caller typically passes a fresh rand.New(rand.NewPCG(...)) seeded
// from the current time.
//
// If the number of players is odd, team A receives one extra player.
type RandomAssignmentStrategy struct {
	rng *rand.Rand
}

// NewRandomAssignmentStrategy builds a RandomAssignmentStrategy with the
// given random source. The rng parameter must not be nil.
func NewRandomAssignmentStrategy(rng *rand.Rand) (*RandomAssignmentStrategy, error) {
	if rng == nil {
		return nil, fmt.Errorf("%w: random source must not be nil", domainerrors.ErrInvalidAssignment)
	}
	return &RandomAssignmentStrategy{rng: rng}, nil
}

// Assign shuffles the players and splits them in half. Team A gets the
// first half (with one extra player on odd counts), team B gets the rest.
func (s *RandomAssignmentStrategy) Assign(players []*entities.Player) ([]entities.PlayerID, []entities.PlayerID, error) {
	if len(players) < minPlayersForAssignment {
		return nil, nil, fmt.Errorf("%w: need at least %d players, got %d", domainerrors.ErrInvalidAssignment, minPlayersForAssignment, len(players))
	}

	// Build a slice of IDs we can shuffle without mutating the input.
	shuffled := make([]entities.PlayerID, len(players))
	for index, player := range players {
		shuffled[index] = player.ID()
	}

	s.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Split point: ceil(n/2) so team A gets the extra on odd counts.
	splitIndex := (len(shuffled) + 1) / 2
	teamA := shuffled[:splitIndex]
	teamB := shuffled[splitIndex:]

	return teamA, teamB, nil
}
