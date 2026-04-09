package entities

import (
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const (
	minVoteScore = 1
	maxVoteScore = 5
)

// VoteID is the unique identifier of a Vote.
type VoteID string

// Vote represents a post-match peer rating: a player rates one of their
// teammates on a scale from 1 to 5.
//
// A vote is always between two players who were on the same team during
// the same match. The "same team" constraint is enforced by the use case
// that creates the vote, not by this entity.
type Vote struct {
	id        VoteID
	matchID   MatchID
	voterID   PlayerID
	votedID   PlayerID
	score     int
	createdAt time.Time
}

// NewVote builds a Vote after validating its inputs.
//
// The score must be between MinVoteScore and MaxVoteScore (inclusive).
// A player cannot vote for themselves; this is enforced here as a basic
// safety check, but the use case should also validate it earlier with
// a more contextual error.
func NewVote(
	id VoteID,
	matchID MatchID,
	voterID PlayerID,
	votedID PlayerID,
	score int,
	createdAt time.Time,
) (*Vote, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if matchID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if voterID == "" || votedID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if voterID == votedID {
		return nil, domainerrors.ErrSelfVote
	}

	if score < minVoteScore || score > maxVoteScore {
		return nil, domainerrors.ErrInvalidScore
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	return &Vote{
		id:        id,
		matchID:   matchID,
		voterID:   voterID,
		votedID:   votedID,
		score:     score,
		createdAt: createdAt,
	}, nil
}

// MinVoteScore returns the minimum allowed score for a vote.
func MinVoteScore() int { return minVoteScore }

// MaxVoteScore returns the maximum allowed score for a vote.
func MaxVoteScore() int { return maxVoteScore }

// ID returns the vote identifier.
func (v *Vote) ID() VoteID { return v.id }

// MatchID returns the identifier of the match this vote belongs to.
func (v *Vote) MatchID() MatchID { return v.matchID }

// VoterID returns the identifier of the player who cast the vote.
func (v *Vote) VoterID() PlayerID { return v.voterID }

// VotedID returns the identifier of the player who received the vote.
func (v *Vote) VotedID() PlayerID { return v.votedID }

// Score returns the vote score on the [MinVoteScore, MaxVoteScore] scale.
func (v *Vote) Score() int { return v.score }

// CreatedAt returns the creation timestamp of the vote.
func (v *Vote) CreatedAt() time.Time { return v.createdAt }
