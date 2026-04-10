package mappers

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// GroupToDomain converts a sqlc-generated Groups row into a domain Group
// entity. Returns an error if the row contains data that fails the
// domain validation (e.g., empty name, zero timestamp), which would
// indicate a corrupted database row.
//
// Going through the NewGroup constructor provides defense in depth: even
// if invalid data ever ends up in the database (manual SQL, broken
// migration), the domain refuses to materialize it.
func GroupToDomain(row generated.Groups) (*entities.Group, error) {
	return entities.NewGroup(
		entities.GroupID(row.ID),
		row.Name,
		entities.OrganizerID(row.OrganizerID),
		row.CreatedAt.Time,
	)
}

// GroupToCreateParams converts a domain Group entity into the parameter
// struct expected by the generated CreateGroup function.
func GroupToCreateParams(group *entities.Group) generated.CreateGroupParams {
	return generated.CreateGroupParams{
		ID:          string(group.ID()),
		OrganizerID: string(group.OrganizerID()),
		Name:        group.Name(),
		CreatedAt:   pgtype.Timestamptz{Time: group.CreatedAt(), Valid: true},
	}
}

// GroupToUpdateParams converts a domain Group entity into the parameter
// struct expected by the generated UpdateGroup function.
//
// Only the fields that the UpdateGroup query touches are populated:
// the ID (for the WHERE clause) and the name (for the SET clause).
func GroupToUpdateParams(group *entities.Group) generated.UpdateGroupParams {
	return generated.UpdateGroupParams{
		ID:   string(group.ID()),
		Name: group.Name(),
	}
}
