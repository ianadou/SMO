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

// PostgresMatchRepository is the PostgreSQL implementation of the
// domain MatchRepository port.
type PostgresMatchRepository struct {
	queries *generated.Queries
}

// NewPostgresMatchRepository builds a PostgresMatchRepository on top of
// the given DBTX.
func NewPostgresMatchRepository(db generated.DBTX) *PostgresMatchRepository {
	return &PostgresMatchRepository{queries: generated.New(db)}
}

// Save persists a new match. Translates foreign key violations into
// ErrReferencedEntityNotFound so the caller does not need to know about pgx.
func (r *PostgresMatchRepository) Save(ctx context.Context, match *entities.Match) error {
	params := mappers.MatchToCreateParams(match)
	if _, err := r.queries.CreateMatch(ctx, params); err != nil {
		if isMatchForeignKeyViolation(err) {
			return fmt.Errorf(
				"postgres match repository: save match %q: %w",
				match.ID(), domainerrors.ErrReferencedEntityNotFound,
			)
		}
		return fmt.Errorf("postgres match repository: save match %q: %w", match.ID(), err)
	}
	return nil
}

// FindByID looks up a match by its identifier.
func (r *PostgresMatchRepository) FindByID(ctx context.Context, id entities.MatchID) (*entities.Match, error) {
	row, err := r.queries.GetMatchByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres match repository: find match %q: %w", id, domainerrors.ErrMatchNotFound)
		}
		return nil, fmt.Errorf("postgres match repository: find match %q: %w", id, err)
	}

	match, err := mappers.MatchToDomain(row)
	if err != nil {
		return nil, fmt.Errorf("postgres match repository: map match %q to domain: %w", id, err)
	}
	return match, nil
}

// ListByGroup returns all matches in the given group, newest first.
func (r *PostgresMatchRepository) ListByGroup(ctx context.Context, groupID entities.GroupID) ([]*entities.Match, error) {
	rows, err := r.queries.ListMatchesByGroupID(ctx, string(groupID))
	if err != nil {
		return nil, fmt.Errorf("postgres match repository: list matches by group %q: %w", groupID, err)
	}

	matches := make([]*entities.Match, 0, len(rows))
	for _, row := range rows {
		match, mapErr := mappers.MatchToDomain(row)
		if mapErr != nil {
			return nil, fmt.Errorf("postgres match repository: map match %q to domain: %w", row.ID, mapErr)
		}
		matches = append(matches, match)
	}
	return matches, nil
}

// UpdateStatus persists a new status for the given match.
func (r *PostgresMatchRepository) UpdateStatus(ctx context.Context, match *entities.Match) error {
	params := mappers.MatchToUpdateStatusParams(match)
	if _, err := r.queries.UpdateMatchStatus(ctx, params); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("postgres match repository: update status match %q: %w", match.ID(), domainerrors.ErrMatchNotFound)
		}
		return fmt.Errorf("postgres match repository: update status match %q: %w", match.ID(), err)
	}
	return nil
}

// Delete removes a match by its identifier. Idempotent.
func (r *PostgresMatchRepository) Delete(ctx context.Context, id entities.MatchID) error {
	if err := r.queries.DeleteMatch(ctx, string(id)); err != nil {
		return fmt.Errorf("postgres match repository: delete match %q: %w", id, err)
	}
	return nil
}

// isMatchForeignKeyViolation returns true if the given error wraps a
// PostgreSQL foreign key violation. Same pattern as the group repository.
func isMatchForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.ForeignKeyViolation
}
