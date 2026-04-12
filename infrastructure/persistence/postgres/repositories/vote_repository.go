package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/mappers"
)

// PostgresVoteRepository is the PostgreSQL implementation of VoteRepository.
type PostgresVoteRepository struct {
	queries *generated.Queries
}

// NewPostgresVoteRepository builds the repository.
func NewPostgresVoteRepository(db generated.DBTX) *PostgresVoteRepository {
	return &PostgresVoteRepository{queries: generated.New(db)}
}

// Save persists a new vote. Translates the FK and UNIQUE violations
// into domain errors.
func (r *PostgresVoteRepository) Save(ctx context.Context, vote *entities.Vote) error {
	params := mappers.VoteToCreateParams(vote)
	if _, err := r.queries.CreateVote(ctx, params); err != nil {
		if isVoteUniqueViolation(err) {
			return fmt.Errorf("postgres vote repository: save %q: %w",
				vote.ID(), domainerrors.ErrAlreadyVoted)
		}
		if isVoteForeignKeyViolation(err) {
			return fmt.Errorf("postgres vote repository: save %q: %w",
				vote.ID(), domainerrors.ErrReferencedEntityNotFound)
		}
		return fmt.Errorf("postgres vote repository: save %q: %w", vote.ID(), err)
	}
	return nil
}

// FindByID looks up a vote by its identifier.
func (r *PostgresVoteRepository) FindByID(ctx context.Context, id entities.VoteID) (*entities.Vote, error) {
	row, err := r.queries.GetVoteByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres vote repository: find %q: %w",
				id, domainerrors.ErrVoteNotFound)
		}
		return nil, fmt.Errorf("postgres vote repository: find %q: %w", id, err)
	}
	vote, mapErr := mappers.VoteToDomain(row)
	if mapErr != nil {
		return nil, fmt.Errorf("postgres vote repository: map %q: %w", id, mapErr)
	}
	return vote, nil
}

// ListByMatch returns votes for a match, oldest first (order of casting).
func (r *PostgresVoteRepository) ListByMatch(ctx context.Context, matchID entities.MatchID) ([]*entities.Vote, error) {
	rows, err := r.queries.ListVotesByMatchID(ctx, string(matchID))
	if err != nil {
		return nil, fmt.Errorf("postgres vote repository: list by match %q: %w", matchID, err)
	}
	return voteRowsToDomain(rows)
}

// ListByVoter returns votes cast by a voter, newest first.
func (r *PostgresVoteRepository) ListByVoter(ctx context.Context, voterID entities.PlayerID) ([]*entities.Vote, error) {
	rows, err := r.queries.ListVotesByVoterID(ctx, string(voterID))
	if err != nil {
		return nil, fmt.Errorf("postgres vote repository: list by voter %q: %w", voterID, err)
	}
	return voteRowsToDomain(rows)
}

// Delete removes a vote by identifier.
func (r *PostgresVoteRepository) Delete(ctx context.Context, id entities.VoteID) error {
	if err := r.queries.DeleteVote(ctx, string(id)); err != nil {
		return fmt.Errorf("postgres vote repository: delete %q: %w", id, err)
	}
	return nil
}

func voteRowsToDomain(rows []generated.Votes) ([]*entities.Vote, error) {
	votes := make([]*entities.Vote, 0, len(rows))
	for _, row := range rows {
		v, err := mappers.VoteToDomain(row)
		if err != nil {
			return nil, fmt.Errorf("map vote %q: %w", row.ID, err)
		}
		votes = append(votes, v)
	}
	return votes, nil
}

func isVoteForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.ForeignKeyViolation
}

func isVoteUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.UniqueViolation
}
