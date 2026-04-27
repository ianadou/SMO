package mappers

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// OrganizerToDomain converts a sqlc-generated Organizers row into a
// domain Organizer entity.
func OrganizerToDomain(row generated.Organizers) (*entities.Organizer, error) {
	return entities.NewOrganizer(
		entities.OrganizerID(row.ID),
		row.Email,
		row.DisplayName,
		row.PasswordHash,
		row.CreatedAt.Time,
	)
}

// OrganizerToCreateParams converts a domain Organizer into the
// parameter struct expected by CreateOrganizer.
func OrganizerToCreateParams(o *entities.Organizer) generated.CreateOrganizerParams {
	return generated.CreateOrganizerParams{
		ID:           string(o.ID()),
		Email:        o.Email(),
		PasswordHash: o.PasswordHash(),
		DisplayName:  o.DisplayName(),
		CreatedAt:    pgtype.Timestamptz{Time: o.CreatedAt(), Valid: true},
	}
}
