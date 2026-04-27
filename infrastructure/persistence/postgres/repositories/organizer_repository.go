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

// PostgresOrganizerRepository is the PostgreSQL implementation of the
// OrganizerRepository port.
type PostgresOrganizerRepository struct {
	queries *generated.Queries
}

// NewPostgresOrganizerRepository builds a PostgresOrganizerRepository.
func NewPostgresOrganizerRepository(db generated.DBTX) *PostgresOrganizerRepository {
	return &PostgresOrganizerRepository{queries: generated.New(db)}
}

// Save persists a new organizer. Translates a UNIQUE violation on
// email into ErrEmailAlreadyExists.
func (r *PostgresOrganizerRepository) Save(ctx context.Context, organizer *entities.Organizer) error {
	params := mappers.OrganizerToCreateParams(organizer)
	if _, err := r.queries.CreateOrganizer(ctx, params); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("postgres organizer repository: save %q: %w",
				organizer.ID(), domainerrors.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("postgres organizer repository: save %q: %w", organizer.ID(), err)
	}
	return nil
}

// FindByID looks up an organizer by ID.
func (r *PostgresOrganizerRepository) FindByID(ctx context.Context, id entities.OrganizerID) (*entities.Organizer, error) {
	row, err := r.queries.GetOrganizerByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres organizer repository: find %q: %w",
				id, domainerrors.ErrOrganizerNotFound)
		}
		return nil, fmt.Errorf("postgres organizer repository: find %q: %w", id, err)
	}
	return mappers.OrganizerToDomain(row)
}

// FindByEmail looks up an organizer by email.
func (r *PostgresOrganizerRepository) FindByEmail(ctx context.Context, email string) (*entities.Organizer, error) {
	row, err := r.queries.GetOrganizerByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres organizer repository: find by email: %w",
				domainerrors.ErrOrganizerNotFound)
		}
		return nil, fmt.Errorf("postgres organizer repository: find by email: %w", err)
	}
	return mappers.OrganizerToDomain(row)
}

// isUniqueViolation reports whether the error wraps a Postgres unique
// constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.UniqueViolation
}
