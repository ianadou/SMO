package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/mappers"
)

// PostgresPlayerRepository is the PostgreSQL implementation of PlayerRepository.
type PostgresPlayerRepository struct {
	queries *generated.Queries
}

// NewPostgresPlayerRepository builds the repository.
func NewPostgresPlayerRepository(db generated.DBTX) *PostgresPlayerRepository {
	return &PostgresPlayerRepository{queries: generated.New(db)}
}

// Save persists a new player. Foreign key violations on group_id are
// translated into ErrReferencedEntityNotFound.
func (r *PostgresPlayerRepository) Save(ctx context.Context, player *entities.Player) error {
	params := mappers.PlayerToCreateParams(player, time.Now())
	if _, err := r.queries.CreatePlayer(ctx, params); err != nil {
		if isPlayerForeignKeyViolation(err) {
			return fmt.Errorf("postgres player repository: save player %q: %w",
				player.ID(), domainerrors.ErrReferencedEntityNotFound)
		}
		return fmt.Errorf("postgres player repository: save player %q: %w", player.ID(), err)
	}
	return nil
}

// FindByID looks up a player by identifier.
func (r *PostgresPlayerRepository) FindByID(ctx context.Context, id entities.PlayerID) (*entities.Player, error) {
	row, err := r.queries.GetPlayerByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres player repository: find player %q: %w",
				id, domainerrors.ErrPlayerNotFound)
		}
		return nil, fmt.Errorf("postgres player repository: find player %q: %w", id, err)
	}
	player, err := mappers.PlayerToDomain(row)
	if err != nil {
		return nil, fmt.Errorf("postgres player repository: map player %q: %w", id, err)
	}
	return player, nil
}

// ListByGroup returns all players in the given group, ordered by ranking DESC.
func (r *PostgresPlayerRepository) ListByGroup(ctx context.Context, groupID entities.GroupID) ([]*entities.Player, error) {
	rows, err := r.queries.ListPlayersByGroupID(ctx, string(groupID))
	if err != nil {
		return nil, fmt.Errorf("postgres player repository: list players by group %q: %w", groupID, err)
	}
	players := make([]*entities.Player, 0, len(rows))
	for _, row := range rows {
		p, mapErr := mappers.PlayerToDomain(row)
		if mapErr != nil {
			return nil, fmt.Errorf("postgres player repository: map player %q: %w", row.ID, mapErr)
		}
		players = append(players, p)
	}
	return players, nil
}

// UpdateRanking persists a new ranking for the given player.
func (r *PostgresPlayerRepository) UpdateRanking(ctx context.Context, player *entities.Player) error {
	params := mappers.PlayerToUpdateRankingParams(player)
	if _, err := r.queries.UpdatePlayerRanking(ctx, params); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("postgres player repository: update ranking %q: %w",
				player.ID(), domainerrors.ErrPlayerNotFound)
		}
		return fmt.Errorf("postgres player repository: update ranking %q: %w", player.ID(), err)
	}
	return nil
}

// Delete removes a player by identifier.
func (r *PostgresPlayerRepository) Delete(ctx context.Context, id entities.PlayerID) error {
	if err := r.queries.DeletePlayer(ctx, string(id)); err != nil {
		return fmt.Errorf("postgres player repository: delete player %q: %w", id, err)
	}
	return nil
}

// isPlayerForeignKeyViolation checks if the error is a pg FK violation.
func isPlayerForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.ForeignKeyViolation
}
