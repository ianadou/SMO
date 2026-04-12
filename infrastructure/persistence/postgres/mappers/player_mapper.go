package mappers

import (
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// PlayerToDomain converts a sqlc-generated Players row into a domain
// Player entity, going through NewPlayer for defense in depth.
//
// Note: the Player entity does not track createdAt, so the row's
// created_at column is ignored here. It is set by the database default
// at insert time and only used for admin/audit queries.
func PlayerToDomain(row generated.Players) (*entities.Player, error) {
	return entities.NewPlayer(
		entities.PlayerID(row.ID),
		entities.GroupID(row.GroupID),
		row.Name,
		int(row.Ranking),
	)
}

// PlayerToCreateParams converts a domain Player into the parameter
// struct for the generated CreatePlayer function.
func PlayerToCreateParams(player *entities.Player, createdAt time.Time) generated.CreatePlayerParams {
	return generated.CreatePlayerParams{
		ID:        string(player.ID()),
		GroupID:   string(player.GroupID()),
		Name:      player.Name(),
		Ranking:   rankingToInt32(player.Ranking()),
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
}

// PlayerToUpdateRankingParams converts a domain Player into the
// parameter struct for UpdatePlayerRanking.
func PlayerToUpdateRankingParams(player *entities.Player) generated.UpdatePlayerRankingParams {
	return generated.UpdatePlayerRankingParams{
		ID:      string(player.ID()),
		Ranking: rankingToInt32(player.Ranking()),
	}
}

// rankingToInt32 converts a player ranking (int) to the int32 the DB
// schema uses. The domain guarantees rankings stay within reasonable
// bounds (typical Elo range is [100, 3000]) but we clamp defensively
// to prevent silent overflow if a future change produces an out-of-range
// value.
func rankingToInt32(ranking int) int32 {
	if ranking > math.MaxInt32 {
		return math.MaxInt32
	}
	if ranking < math.MinInt32 {
		return math.MinInt32
	}
	return int32(ranking)
}
