package mappers

import (
	"math"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// VoteToDomain converts a sqlc-generated Votes row into a domain Vote.
func VoteToDomain(row generated.Votes) (*entities.Vote, error) {
	return entities.NewVote(
		entities.VoteID(row.ID),
		entities.MatchID(row.MatchID),
		entities.PlayerID(row.VoterID),
		entities.PlayerID(row.VotedID),
		int(row.Score),
		row.CreatedAt.Time,
	)
}

// VoteToCreateParams converts a domain Vote into the parameter struct
// for the generated CreateVote function.
func VoteToCreateParams(vote *entities.Vote) generated.CreateVoteParams {
	return generated.CreateVoteParams{
		ID:        string(vote.ID()),
		MatchID:   string(vote.MatchID()),
		VoterID:   string(vote.VoterID()),
		VotedID:   string(vote.VotedID()),
		Score:     scoreToInt32(vote.Score()),
		CreatedAt: pgtype.Timestamptz{Time: vote.CreatedAt(), Valid: true},
	}
}

// scoreToInt32 converts the domain vote score (int) to int32 used by
// the DB schema. The domain guarantees the score is in [1, 5] but we
// clamp defensively to satisfy the gosec G115 linter.
func scoreToInt32(score int) int32 {
	if score > math.MaxInt32 {
		return math.MaxInt32
	}
	if score < math.MinInt32 {
		return math.MinInt32
	}
	return int32(score)
}
