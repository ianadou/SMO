package ports

import (
	"context"

	"github.com/ianadou/smo/domain/entities"
)

// VoteRepository is the persistence contract for the Vote aggregate.
//
// Votes are immutable once cast; there is no Update method. A vote can
// be Deleted (e.g., to correct an obvious mistake), but mutating the
// score would violate the audit semantics of a peer rating.
type VoteRepository interface {
	// Save persists a new vote. Returns ErrAlreadyVoted (wrapped) if the
	// unique constraint (match_id, voter_id, voted_id) is violated, or
	// ErrReferencedEntityNotFound if any FK is invalid.
	Save(ctx context.Context, vote *entities.Vote) error

	// FindByID looks up a vote by its identifier.
	FindByID(ctx context.Context, id entities.VoteID) (*entities.Vote, error)

	// ListByMatch returns all votes cast during the given match.
	ListByMatch(ctx context.Context, matchID entities.MatchID) ([]*entities.Vote, error)

	// ListByVoter returns all votes cast by the given voter, across
	// all matches.
	ListByVoter(ctx context.Context, voterID entities.PlayerID) ([]*entities.Vote, error)

	// Delete removes a vote by identifier.
	Delete(ctx context.Context, id entities.VoteID) error
}
