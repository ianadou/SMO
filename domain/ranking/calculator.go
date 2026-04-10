package ranking

import (
	"fmt"
	"math"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

const (
	// defaultLearningRate is the default blending factor between the
	// previous ranking and the new average score from votes. A value of
	// 0.1 means a perfect score of 5/5 moves the ranking by +20 points
	// and a worst score of 1/5 moves it by -20 points.
	defaultLearningRate = 0.1

	// neutralVoteScore is the score that produces zero ranking change.
	// Votes above neutral push the ranking up, votes below push it down.
	neutralVoteScore = 3.0

	// scoreNormalizationFactor scales a centered score (in [-2, +2]) to a
	// raw ranking adjustment (in [-200, +200]) before applying the
	// learning rate.
	scoreNormalizationFactor = 100.0

	// minRanking is the lower bound of a player's ranking after update.
	// Rankings cannot go below zero.
	minRanking = 0
)

// Calculator computes updated rankings from a set of votes received
// during a match.
type Calculator struct {
	learningRate float64
}

// NewCalculator builds a Calculator with the given learning rate.
//
// The learning rate must be in the half-open interval (0, 1]. A value
// close to 0 means rankings change slowly (stable), a value close to 1
// means rankings change quickly (reactive).
//
// Returns ErrInvalidParameter if the learning rate is out of range.
func NewCalculator(learningRate float64) (*Calculator, error) {
	if learningRate <= 0 || learningRate > 1 {
		return nil, fmt.Errorf("%w: learning rate must be in (0, 1], got %v", domainerrors.ErrInvalidParameter, learningRate)
	}
	return &Calculator{learningRate: learningRate}, nil
}

// DefaultLearningRate returns the default blending factor used by the
// calculator when callers do not have a specific value in mind.
func DefaultLearningRate() float64 { return defaultLearningRate }

// Compute returns the updated ranking for each player based on the votes
// they received. Players who received no vote keep their current ranking
// in the result map.
//
// Returns an error if a vote references a player that is not in the
// players list.
func (c *Calculator) Compute(
	players []*entities.Player,
	votes []*entities.Vote,
) (map[entities.PlayerID]int, error) {
	if len(players) == 0 {
		return map[entities.PlayerID]int{}, nil
	}

	// Index players by ID for O(1) lookup and to detect votes for
	// players that are not in the input list.
	playerByID := make(map[entities.PlayerID]*entities.Player, len(players))
	for _, player := range players {
		playerByID[player.ID()] = player
	}

	// Group votes by the player they target.
	votesReceivedBy := make(map[entities.PlayerID][]*entities.Vote, len(players))
	for _, vote := range votes {
		if _, exists := playerByID[vote.VotedID()]; !exists {
			return nil, fmt.Errorf("%w: vote targets player %q not in input list", domainerrors.ErrInvalidParameter, vote.VotedID())
		}
		votesReceivedBy[vote.VotedID()] = append(votesReceivedBy[vote.VotedID()], vote)
	}

	updated := make(map[entities.PlayerID]int, len(players))
	for _, player := range players {
		votesForPlayer := votesReceivedBy[player.ID()]
		updated[player.ID()] = c.computeNewRanking(player.Ranking(), votesForPlayer)
	}

	return updated, nil
}

// computeNewRanking applies the weighted-average formula to a single
// player given their current ranking and the votes they received.
//
// If no votes were received, the ranking is returned unchanged.
func (c *Calculator) computeNewRanking(currentRanking int, votes []*entities.Vote) int {
	if len(votes) == 0 {
		return currentRanking
	}

	totalScore := 0
	for _, vote := range votes {
		totalScore += vote.Score()
	}
	averageScore := float64(totalScore) / float64(len(votes))

	centeredScore := averageScore - neutralVoteScore
	rawAdjustment := centeredScore * scoreNormalizationFactor
	adjustment := rawAdjustment * c.learningRate

	newRanking := currentRanking + int(math.Round(adjustment))
	if newRanking < minRanking {
		newRanking = minRanking
	}
	return newRanking
}
