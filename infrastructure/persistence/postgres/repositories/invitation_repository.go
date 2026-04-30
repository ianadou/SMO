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

// PostgresInvitationRepository is the PostgreSQL implementation of
// InvitationRepository.
type PostgresInvitationRepository struct {
	queries *generated.Queries
}

// NewPostgresInvitationRepository builds the repository.
func NewPostgresInvitationRepository(db generated.DBTX) *PostgresInvitationRepository {
	return &PostgresInvitationRepository{queries: generated.New(db)}
}

// Save persists a new invitation. FK violations on match_id become
// ErrReferencedEntityNotFound.
func (r *PostgresInvitationRepository) Save(ctx context.Context, inv *entities.Invitation) error {
	params := mappers.InvitationToCreateParams(inv)
	if _, err := r.queries.CreateInvitation(ctx, params); err != nil {
		if isInvitationForeignKeyViolation(err) {
			return fmt.Errorf("postgres invitation repository: save %q: %w",
				inv.ID(), domainerrors.ErrReferencedEntityNotFound)
		}
		return fmt.Errorf("postgres invitation repository: save %q: %w", inv.ID(), err)
	}
	return nil
}

// FindByID looks up an invitation by identifier.
func (r *PostgresInvitationRepository) FindByID(ctx context.Context, id entities.InvitationID) (*entities.Invitation, error) {
	row, err := r.queries.GetInvitationByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres invitation repository: find %q: %w",
				id, domainerrors.ErrInvitationNotFound)
		}
		return nil, fmt.Errorf("postgres invitation repository: find %q: %w", id, err)
	}
	inv, mapErr := mappers.InvitationToDomain(row)
	if mapErr != nil {
		return nil, fmt.Errorf("postgres invitation repository: map %q: %w", id, mapErr)
	}
	return inv, nil
}

// FindByTokenHash looks up an invitation by its token hash. Used by
// AcceptInvitationUseCase to find an invitation from a submitted token.
func (r *PostgresInvitationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.Invitation, error) {
	row, err := r.queries.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres invitation repository: find by hash: %w",
				domainerrors.ErrInvitationNotFound)
		}
		return nil, fmt.Errorf("postgres invitation repository: find by hash: %w", err)
	}
	inv, mapErr := mappers.InvitationToDomain(row)
	if mapErr != nil {
		return nil, fmt.Errorf("postgres invitation repository: map by hash: %w", mapErr)
	}
	return inv, nil
}

// ListByMatch returns all invitations for the given match, newest first.
func (r *PostgresInvitationRepository) ListByMatch(ctx context.Context, matchID entities.MatchID) ([]*entities.Invitation, error) {
	rows, err := r.queries.ListInvitationsByMatchID(ctx, string(matchID))
	if err != nil {
		return nil, fmt.Errorf("postgres invitation repository: list by match %q: %w", matchID, err)
	}
	invitations := make([]*entities.Invitation, 0, len(rows))
	for _, row := range rows {
		inv, mapErr := mappers.InvitationToDomain(row)
		if mapErr != nil {
			return nil, fmt.Errorf("postgres invitation repository: map %q: %w", row.ID, mapErr)
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}

// CountConfirmedByMatch returns the count of used (used_at IS NOT NULL)
// invitations for the given match.
func (r *PostgresInvitationRepository) CountConfirmedByMatch(ctx context.Context, matchID entities.MatchID) (int, error) {
	count, err := r.queries.CountConfirmedInvitationsByMatchID(ctx, string(matchID))
	if err != nil {
		return 0, fmt.Errorf("postgres invitation repository: count confirmed for %q: %w", matchID, err)
	}
	return int(count), nil
}

// MarkAsUsed persists the used_at timestamp for the given invitation.
func (r *PostgresInvitationRepository) MarkAsUsed(ctx context.Context, inv *entities.Invitation) error {
	params := mappers.InvitationToMarkAsUsedParams(inv)
	if _, err := r.queries.MarkInvitationAsUsed(ctx, params); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("postgres invitation repository: mark as used %q: %w",
				inv.ID(), domainerrors.ErrInvitationNotFound)
		}
		return fmt.Errorf("postgres invitation repository: mark as used %q: %w", inv.ID(), err)
	}
	return nil
}

// Delete removes an invitation by identifier.
func (r *PostgresInvitationRepository) Delete(ctx context.Context, id entities.InvitationID) error {
	if err := r.queries.DeleteInvitation(ctx, string(id)); err != nil {
		return fmt.Errorf("postgres invitation repository: delete %q: %w", id, err)
	}
	return nil
}

// isInvitationForeignKeyViolation checks if the error is a pg FK violation.
func isInvitationForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.ForeignKeyViolation
}
