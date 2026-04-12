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
	var usedAtPtr *time.Time
	if row.UsedAt.Valid {
		t := row.UsedAt.Time
		usedAtPtr = &t
	}
	return entities.NewInvitation(
		entities.InvitationID(row.ID),
		entities.MatchID(row.MatchID),
		row.TokenHash,
		row.ExpiresAt.Time,
		usedAtPtr,
		row.CreatedAt.Time,
	)
}

// InvitationToCreateParams converts a domain Invitation into create params.
func InvitationToCreateParams(inv *entities.Invitation) generated.CreateInvitationParams {
	return generated.CreateInvitationParams{
		ID:        string(inv.ID()),
		MatchID:   string(inv.MatchID()),
		TokenHash: inv.TokenHash(),
		ExpiresAt: pgtype.Timestamptz{Time: inv.ExpiresAt(), Valid: true},
		UsedAt:    usedAtToPg(inv.UsedAt()),
		CreatedAt: pgtype.Timestamptz{Time: inv.CreatedAt(), Valid: true},
	}
}

// InvitationToMarkAsUsedParams converts a domain Invitation into mark-as-used params.
func InvitationToMarkAsUsedParams(inv *entities.Invitation) generated.MarkInvitationAsUsedParams {
	return generated.MarkInvitationAsUsedParams{
		ID:     string(inv.ID()),
		UsedAt: usedAtToPg(inv.UsedAt()),
	}
}

// usedAtToPg converts an optional *time.Time into pgtype.Timestamptz.
// Nil becomes the NULL representation.
func usedAtToPg(usedAt *time.Time) pgtype.Timestamptz {
	if usedAt == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *usedAt, Valid: true}
}
