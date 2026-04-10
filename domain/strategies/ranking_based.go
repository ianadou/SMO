package strategies

import (
	"fmt"
	"sort"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// RankingBasedStrategy distributes players into two teams using a snake
// draft based on their ranking, so that both teams end up with a similar
// total ranking.
//
// The algorithm:
//  1. Sort players by ranking, descending (best first).
//  2. Distribute in snake pairs: each pair takes one player per team,
//     and the team that picks first alternates between pairs.
//
// This guarantees that the top players are split evenly between teams,
// avoiding the case where the top half of the ranking ends up on one side.
type RankingBasedStrategy struct{}

// NewRankingBasedStrategy builds a RankingBasedStrategy. It takes no
// configuration: the algorithm is fully deterministic given its input.
func NewRankingBasedStrategy() *RankingBasedStrategy {
	return &RankingBasedStrategy{}
}

// Assign sorts the players by ranking descending and distributes them
// using a snake draft. If the number of players is odd, team A receives
// one extra player.
func (s *RankingBasedStrategy) Assign(players []*entities.Player) ([]entities.PlayerID, []entities.PlayerID, error) {
	if len(players) < minPlayersForAssignment {
		return nil, nil, fmt.Errorf("%w: need at least %d players, got %d", domainerrors.ErrInvalidAssignment, minPlayersForAssignment, len(players))
	}

	// Defensive copy so we don't mutate the caller's slice with sort.
	sorted := make([]*entities.Player, len(players))
	copy(sorted, players)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Ranking() > sorted[j].Ranking()
	})

	teamA := make([]entities.PlayerID, 0, len(sorted)/2+1)
	teamB := make([]entities.PlayerID, 0, len(sorted)/2+1)

	// Snake draft: each pair of players is split between the two teams,
	// and the team that takes the first slot alternates between pairs.
	for index, player := range sorted {
		pairIndex := index / 2
		positionInPair := index % 2
		startsWithA := pairIndex%2 == 0

		assignToA := (startsWithA && positionInPair == 0) || (!startsWithA && positionInPair == 1)
		if assignToA {
			teamA = append(teamA, player.ID())
		} else {
			teamB = append(teamB, player.ID())
		}
	}

	return teamA, teamB, nil
}
