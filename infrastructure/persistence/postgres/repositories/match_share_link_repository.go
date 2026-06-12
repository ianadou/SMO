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

const matchShareLinkRepoOp = "postgres match share link repository"

// PostgresMatchShareLinkRepository is the PostgreSQL implementation of
// MatchShareLinkRepository.
type PostgresMatchShareLinkRepository struct {
	queries *generated.Queries
}

// NewPostgresMatchShareLinkRepository builds the repository.
func NewPostgresMatchShareLinkRepository(db generated.DBTX) *PostgresMatchShareLinkRepository {
	return &PostgresMatchShareLinkRepository{queries: generated.New(db)}
}

// Create persists a new share link. FK violations on match_id become
// ErrReferencedEntityNotFound.
func (r *PostgresMatchShareLinkRepository) Create(ctx context.Context, link *entities.MatchShareLink) error {
	params := mappers.MatchShareLinkToCreateParams(link)
	if _, err := r.queries.CreateMatchShareLink(ctx, params); err != nil {
		if isMatchShareLinkForeignKeyViolation(err) {
			return fmt.Errorf(matchShareLinkRepoOp+": create %q: %w",
				link.ID(), domainerrors.ErrReferencedEntityNotFound)
		}
		return fmt.Errorf(matchShareLinkRepoOp+": create %q: %w", link.ID(), err)
	}
	return nil
}

// FindByTokenHash looks up a share link by its token hash, regardless of
// its active state: the use case decides what an inactive link means.
func (r *PostgresMatchShareLinkRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.MatchShareLink, error) {
	row, err := r.queries.GetMatchShareLinkByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(matchShareLinkRepoOp+": find by hash: %w",
				domainerrors.ErrShareLinkNotFound)
		}
		return nil, fmt.Errorf(matchShareLinkRepoOp+": find by hash: %w", err)
	}
	link, mapErr := mappers.MatchShareLinkToDomain(row)
	if mapErr != nil {
		return nil, fmt.Errorf(matchShareLinkRepoOp+": map by hash: %w", mapErr)
	}
	return link, nil
}

// FindActiveByMatchID returns the single non-revoked, non-expired link
// of the match. Expiry is evaluated against the database clock (the
// same NOW() that timestamps the rows), keeping the comparison
// consistent with what was stored.
func (r *PostgresMatchShareLinkRepository) FindActiveByMatchID(ctx context.Context, matchID entities.MatchID) (*entities.MatchShareLink, error) {
	row, err := r.queries.GetActiveMatchShareLinkByMatchID(ctx, string(matchID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(matchShareLinkRepoOp+": find active for %q: %w",
				matchID, domainerrors.ErrShareLinkNotFound)
		}
		return nil, fmt.Errorf(matchShareLinkRepoOp+": find active for %q: %w", matchID, err)
	}
	link, mapErr := mappers.MatchShareLinkToDomain(row)
	if mapErr != nil {
		return nil, fmt.Errorf(matchShareLinkRepoOp+": map active for %q: %w", matchID, mapErr)
	}
	return link, nil
}

// Update persists the mutable state of the link (its revocation).
func (r *PostgresMatchShareLinkRepository) Update(ctx context.Context, link *entities.MatchShareLink) error {
	if _, err := r.queries.UpdateMatchShareLink(ctx, mappers.MatchShareLinkToUpdateParams(link)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(matchShareLinkRepoOp+": update %q: %w",
				link.ID(), domainerrors.ErrShareLinkNotFound)
		}
		return fmt.Errorf(matchShareLinkRepoOp+": update %q: %w", link.ID(), err)
	}
	return nil
}

// isMatchShareLinkForeignKeyViolation checks if the error is a pg FK violation.
func isMatchShareLinkForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.ForeignKeyViolation
}
