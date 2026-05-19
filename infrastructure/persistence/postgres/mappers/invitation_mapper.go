package mappers

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// InvitationToDomain converts a sqlc-generated Invitations row into a
// domain Invitation entity.
func InvitationToDomain(row generated.Invitations) (*entities.Invitation, error) {
	var respondedAtPtr *time.Time
	if row.RespondedAt.Valid {
		t := row.RespondedAt.Time
		respondedAtPtr = &t
	}
	return entities.NewInvitation(
		entities.InvitationID(row.ID),
		entities.MatchID(row.MatchID),
		entities.PlayerID(row.PlayerID),
		row.TokenHash,
		row.ExpiresAt.Time,
		entities.InvitationResponse(row.Response),
		respondedAtPtr,
		row.CreatedAt.Time,
	)
}

// InvitationToCreateParams converts a domain Invitation into create
// params. New invitations are always created pending (response and
// responded_at are fixed in the SQL), so only the immutable identity
// fields are mapped here.
func InvitationToCreateParams(inv *entities.Invitation) generated.CreateInvitationParams {
	return generated.CreateInvitationParams{
		ID:        string(inv.ID()),
		MatchID:   string(inv.MatchID()),
		PlayerID:  string(inv.PlayerID()),
		TokenHash: inv.TokenHash(),
		ExpiresAt: pgtype.Timestamptz{Time: inv.ExpiresAt(), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: inv.CreatedAt(), Valid: true},
	}
}

// InvitationToUpdateResponseParams converts a domain Invitation into the
// params for persisting its current response.
func InvitationToUpdateResponseParams(inv *entities.Invitation) generated.UpdateInvitationResponseParams {
	return generated.UpdateInvitationResponseParams{
		ID:          string(inv.ID()),
		Response:    string(inv.Response()),
		RespondedAt: respondedAtToPg(inv.RespondedAt()),
	}
}

// respondedAtToPg converts an optional *time.Time into pgtype.Timestamptz.
// Nil becomes the NULL representation.
func respondedAtToPg(respondedAt *time.Time) pgtype.Timestamptz {
	if respondedAt == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *respondedAt, Valid: true}
}
