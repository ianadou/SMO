package mappers

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// GroupToDomain converts a sqlc-generated Groups row into a domain Group
// entity. Returns an error if the row contains data that fails the
// domain validation (e.g., empty name, zero timestamp, malformed
// webhook URL), which would indicate a corrupted database row.
//
// Going through the NewGroup constructor provides defense in depth: even
// if invalid data ever ends up in the database (manual SQL, broken
// migration), the domain refuses to materialize it.
//
// The webhook column is nullable in Postgres (*string at the sqlc
// boundary). The mapper translates nil to the empty-string convention
// the domain uses, so the rest of the codebase only deals with
// strings.
func GroupToDomain(row generated.Groups) (*entities.Group, error) {
	return entities.NewGroup(
		entities.GroupID(row.ID),
		row.Name,
		entities.OrganizerID(row.OrganizerID),
		webhookFromPointer(row.DiscordWebhookUrl),
		row.CreatedAt.Time,
	)
}

// GroupToCreateParams converts a domain Group entity into the parameter
// struct expected by the generated CreateGroup function.
func GroupToCreateParams(group *entities.Group) generated.CreateGroupParams {
	return generated.CreateGroupParams{
		ID:                string(group.ID()),
		OrganizerID:       string(group.OrganizerID()),
		Name:              group.Name(),
		DiscordWebhookUrl: webhookToPointer(group.WebhookURL()),
		CreatedAt:         pgtype.Timestamptz{Time: group.CreatedAt(), Valid: true},
	}
}

// GroupToUpdateParams converts a domain Group entity into the parameter
// struct expected by the generated UpdateGroup function.
//
// The UpdateGroup query touches name and discord_webhook_url; the ID
// drives the WHERE clause. Other fields (organizer_id, created_at) are
// left untouched at the SQL level.
func GroupToUpdateParams(group *entities.Group) generated.UpdateGroupParams {
	return generated.UpdateGroupParams{
		ID:                string(group.ID()),
		Name:              group.Name(),
		DiscordWebhookUrl: webhookToPointer(group.WebhookURL()),
	}
}

// webhookToPointer converts the domain's empty-string-means-no-webhook
// convention into the *string sqlc expects for nullable columns. An
// empty string maps to nil so the column ends up SQL NULL rather than
// an empty TEXT, which is more honest semantically and what the domain
// expects to read back.
func webhookToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// webhookFromPointer is the inverse of webhookToPointer.
func webhookFromPointer(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
