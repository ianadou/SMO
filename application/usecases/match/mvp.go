package match

import (
	"sort"

	"github.com/ianadou/smo/domain/entities"
)

// selectMVP returns the player elected MVP from the given votes, or nil
// if no votes were cast (no player can claim the title).
//
// The MVP is the player with the highest average vote score. When two
// or more players are tied on average, the tiebreakers are, in order:
//
//  1. Most votes received (a higher sample size beats a lucky 5/5 single)
//  2. Lexicographically smaller PlayerID (deterministic last resort)
//
// The function is pure: same inputs always yield the same output. This
// is what makes the FinalizeMatchUseCase idempotent under retry.
func selectMVP(votes []*entities.Vote) *entities.PlayerID {
	if len(votes) == 0 {
		return nil
	}

	type tally struct {
		playerID entities.PlayerID
		total    int
		count    int
	}
	tallies := make(map[entities.PlayerID]*tally)
	for _, vote := range votes {
		t, ok := tallies[vote.VotedID()]
		if !ok {
			t = &tally{playerID: vote.VotedID()}
			tallies[vote.VotedID()] = t
		}
		t.total += vote.Score()
		t.count++
	}

	candidates := make([]*tally, 0, len(tallies))
	for _, t := range tallies {
		candidates = append(candidates, t)
	}

	sort.Slice(candidates, func(i, j int) bool {
		ai := float64(candidates[i].total) / float64(candidates[i].count)
		aj := float64(candidates[j].total) / float64(candidates[j].count)
		if ai != aj {
			return ai > aj
		}
		if candidates[i].count != candidates[j].count {
			return candidates[i].count > candidates[j].count
		}
		return candidates[i].playerID < candidates[j].playerID
	})

	winner := candidates[0].playerID
	return &winner
}
