package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/mappers"
)

// PostgresGroupRepository is the PostgreSQL implementation of the
// domain GroupRepository port. It uses the sqlc-generated Queries to
// execute SQL and the mappers package to translate between database
// rows and domain entities.
type PostgresGroupRepository struct {
	queries *generated.Queries
}

// NewPostgresGroupRepository builds a PostgresGroupRepository on top of
// the given DBTX. The DBTX can be a *pgx.Conn, a *pgxpool.Pool, or a
// transaction; the repository does not care about the underlying type.
func NewPostgresGroupRepository(db generated.DBTX) *PostgresGroupRepository {
	return &PostgresGroupRepository{queries: generated.New(db)}
}

// Save persists a new group by inserting a row into the groups table.
func (r *PostgresGroupRepository) Save(ctx context.Context, group *entities.Group) error {
	params := mappers.GroupToCreateParams(group)
	if _, err := r.queries.CreateGroup(ctx, params); err != nil {
		return fmt.Errorf("postgres group repository: save group %q: %w", group.ID(), err)
	}
	return nil
}

// FindByID looks up a group by its identifier. Returns ErrGroupNotFound
// if no group exists with that id.
func (r *PostgresGroupRepository) FindByID(ctx context.Context, id entities.GroupID) (*entities.Group, error) {
	row, err := r.queries.GetGroupByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres group repository: find group %q: %w", id, domainerrors.ErrGroupNotFound)
		}
		return nil, fmt.Errorf("postgres group repository: find group %q: %w", id, err)
	}

	group, err := mappers.GroupToDomain(row)
	if err != nil {
		return nil, fmt.Errorf("postgres group repository: map group %q to domain: %w", id, err)
	}
	return group, nil
}

// ListByOrganizer returns all groups owned by the given organizer.
func (r *PostgresGroupRepository) ListByOrganizer(ctx context.Context, organizerID entities.OrganizerID) ([]*entities.Group, error) {
	rows, err := r.queries.ListGroupsByOrganizerID(ctx, string(organizerID))
	if err != nil {
		return nil, fmt.Errorf("postgres group repository: list groups by organizer %q: %w", organizerID, err)
	}

	groups := make([]*entities.Group, 0, len(rows))
	for _, row := range rows {
		group, mapErr := mappers.GroupToDomain(row)
		if mapErr != nil {
			return nil, fmt.Errorf("postgres group repository: map group %q to domain: %w", row.ID, mapErr)
		}
		groups = append(groups, group)
	}
	return groups, nil
}

// Update modifies an existing group. Returns ErrGroupNotFound if no
// group exists with that id.
func (r *PostgresGroupRepository) Update(ctx context.Context, group *entities.Group) error {
	params := mappers.GroupToUpdateParams(group)
	if _, err := r.queries.UpdateGroup(ctx, params); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("postgres group repository: update group %q: %w", group.ID(), domainerrors.ErrGroupNotFound)
		}
		return fmt.Errorf("postgres group repository: update group %q: %w", group.ID(), err)
	}
	return nil
}

// Delete removes a group by its identifier. Returns ErrGroupNotFound if
// no group exists with that id.
//
// Note: the generated DeleteGroup query uses :exec which does not signal
// "not found" via pgx.ErrNoRows. To detect a missing row, we would need
// to use :execrows and check the affected count. For now, we accept that
// deleting a non-existent id is a no-op (idempotent), which is a common
// REST DELETE semantic.
func (r *PostgresGroupRepository) Delete(ctx context.Context, id entities.GroupID) error {
	if err := r.queries.DeleteGroup(ctx, string(id)); err != nil {
		return fmt.Errorf("postgres group repository: delete group %q: %w", id, err)
	}
	return nil
}
