package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/mappers"
)

// PostgresMatchRepository is the PostgreSQL implementation of the
// domain MatchRepository port.
//
// It holds the pool (not just a generated.DBTX) because ReplaceTeams
// runs inside an explicit transaction (delete-all then insert), which
// the sqlc DBTX interface does not expose.
type PostgresMatchRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewPostgresMatchRepository builds a PostgresMatchRepository on top of
// the given pool.
func NewPostgresMatchRepository(pool *pgxpool.Pool) *PostgresMatchRepository {
	return &PostgresMatchRepository{pool: pool, queries: generated.New(pool)}
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

	members, err := r.queries.ListMatchTeamMembers(ctx, string(id))
	if err != nil {
		return nil, fmt.Errorf("postgres match repository: load teams %q: %w", id, err)
	}
	teamA, teamB := mappers.TeamsFromMemberRows(members)
	if len(teamA) > 0 || len(teamB) > 0 {
		hydrated, rehErr := entities.RehydrateMatch(entities.MatchSnapshot{
			ID: match.ID(), GroupID: match.GroupID(), Title: match.Title(),
			Venue: match.Venue(), ScheduledAt: match.ScheduledAt(),
			Status: match.Status(), MVPPlayerID: match.MVP(),
			CreatedAt: match.CreatedAt(), TeamA: teamA, TeamB: teamB,
		})
		if rehErr != nil {
			return nil, fmt.Errorf("postgres match repository: rehydrate teams %q: %w", id, rehErr)
		}
		match = hydrated
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

// Finalize persists the MVP and the new status of the given match in
// a single statement. Translates a missing match into ErrMatchNotFound.
func (r *PostgresMatchRepository) Finalize(ctx context.Context, match *entities.Match) error {
	params := mappers.MatchToFinalizeParams(match)
	if _, err := r.queries.FinalizeMatch(ctx, params); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("postgres match repository: finalize match %q: %w", match.ID(), domainerrors.ErrMatchNotFound)
		}
		return fmt.Errorf("postgres match repository: finalize match %q: %w", match.ID(), err)
	}
	return nil
}

// ReplaceTeams atomically replaces the match's team composition.
func (r *PostgresMatchRepository) ReplaceTeams(ctx context.Context, match *entities.Match) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres match repository: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.queries.WithTx(tx)

	if delErr := qtx.DeleteMatchTeamMembers(ctx, string(match.ID())); delErr != nil {
		return fmt.Errorf("postgres match repository: clear teams %q: %w", match.ID(), delErr)
	}
	for _, p := range mappers.MatchTeamMemberInsertParams(match) {
		if insErr := qtx.InsertMatchTeamMember(ctx, p); insErr != nil {
			if isMatchForeignKeyViolation(insErr) {
				return fmt.Errorf("postgres match repository: insert team member %q: %w",
					match.ID(), domainerrors.ErrReferencedEntityNotFound)
			}
			return fmt.Errorf("postgres match repository: insert team member %q: %w", match.ID(), insErr)
		}
	}
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("postgres match repository: commit teams %q: %w", match.ID(), commitErr)
	}
	return nil
}

// ListTeamMembersWithPlayers returns team membership joined with players.
func (r *PostgresMatchRepository) ListTeamMembersWithPlayers(ctx context.Context, matchID entities.MatchID) ([]entities.MatchTeamMember, error) {
	rows, err := r.queries.ListMatchTeamMembersWithPlayers(ctx, string(matchID))
	if err != nil {
		return nil, fmt.Errorf("postgres match repository: list team members %q: %w", matchID, err)
	}
	out := make([]entities.MatchTeamMember, 0, len(rows))
	for _, row := range rows {
		out = append(out, entities.MatchTeamMember{
			PlayerID: entities.PlayerID(row.PlayerID), PlayerName: row.Name,
			Team: row.Team, Slot: int(row.Slot),
		})
	}
	return out, nil
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
