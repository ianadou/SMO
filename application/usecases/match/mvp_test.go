package match

import (
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func voteScored(id entities.VoteID, voter, voted entities.PlayerID, score int) *entities.Vote {
	v, err := entities.NewVote(id, "match-1", voter, voted, score, time.Now())
	if err != nil {
		panic(err)
	}
	return v
}

func TestSelectMVP_ReturnsNil_WhenNoVotes(t *testing.T) {
	t.Parallel()

	mvp := selectMVP(nil)

	if mvp != nil {
		t.Errorf("expected nil MVP, got %v", mvp)
	}
}

func TestSelectMVP_ReturnsHighestAverage(t *testing.T) {
	t.Parallel()

	votes := []*entities.Vote{
		voteScored("v-1", "voter-1", "p-a", 5),
		voteScored("v-2", "voter-2", "p-a", 5),
		voteScored("v-3", "voter-1", "p-b", 3),
		voteScored("v-4", "voter-2", "p-b", 4),
	}

	mvp := selectMVP(votes)

	if mvp == nil || *mvp != "p-a" {
		t.Errorf("expected MVP p-a (avg 5.0), got %v", mvp)
	}
}

func TestSelectMVP_TiebreakerByVoteCount(t *testing.T) {
	t.Parallel()

	// Both p-a and p-b have an average of 5.0. p-b has more votes (3 vs 1)
	// and must therefore win the tiebreaker.
	votes := []*entities.Vote{
		voteScored("v-1", "voter-1", "p-a", 5),
		voteScored("v-2", "voter-1", "p-b", 5),
		voteScored("v-3", "voter-2", "p-b", 5),
		voteScored("v-4", "voter-3", "p-b", 5),
	}

	mvp := selectMVP(votes)

	if mvp == nil || *mvp != "p-b" {
		t.Errorf("expected MVP p-b (more votes at same average), got %v", mvp)
	}
}

func TestSelectMVP_TiebreakerByPlayerID_WhenAverageAndCountAreEqual(t *testing.T) {
	t.Parallel()

	// Two players with identical average AND identical vote count: the
	// lexicographically smaller PlayerID wins (deterministic last resort).
	votes := []*entities.Vote{
		voteScored("v-1", "voter-1", "p-z", 5),
		voteScored("v-2", "voter-1", "p-a", 5),
	}

	mvp := selectMVP(votes)

	if mvp == nil || *mvp != "p-a" {
		t.Errorf("expected MVP p-a (lex smaller), got %v", mvp)
	}
}
