package mappers

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// MatchShareLinkToDomain converts a sqlc-generated MatchShareLinks row
// into a domain MatchShareLink entity.
func MatchShareLinkToDomain(row generated.MatchShareLinks) (*entities.MatchShareLink, error) {
	return entities.NewMatchShareLink(
		entities.MatchShareLinkID(row.ID),
		entities.MatchID(row.MatchID),
		row.TokenHash,
		row.ExpiresAt.Time,
		optionalTimeFromPg(row.RevokedAt),
		row.CreatedAt.Time,
	)
}

// MatchShareLinkToCreateParams converts a domain MatchShareLink into
// create params. New links are always created unrevoked (revoked_at is
// fixed to NULL in the SQL), so only the immutable fields are mapped.
func MatchShareLinkToCreateParams(link *entities.MatchShareLink) generated.CreateMatchShareLinkParams {
	return generated.CreateMatchShareLinkParams{
		ID:        string(link.ID()),
		MatchID:   string(link.MatchID()),
		TokenHash: link.TokenHash(),
		ExpiresAt: pgtype.Timestamptz{Time: link.ExpiresAt(), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: link.CreatedAt(), Valid: true},
	}
}

// MatchShareLinkToUpdateParams converts a domain MatchShareLink into the
// params for persisting its revocation state.
func MatchShareLinkToUpdateParams(link *entities.MatchShareLink) generated.UpdateMatchShareLinkParams {
	return generated.UpdateMatchShareLinkParams{
		ID:        string(link.ID()),
		RevokedAt: optionalTimeToPg(link.RevokedAt()),
	}
}
