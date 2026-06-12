package mappers

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// InvitationToDomain converts a sqlc-generated Invitations row into a
// domain Invitation entity.
func InvitationToDomain(row generated.Invitations) (*entities.Invitation, error) {
	return entities.NewInvitation(
		entities.InvitationID(row.ID),
		entities.MatchID(row.MatchID),
		entities.PlayerID(row.PlayerID),
		row.TokenHash,
		row.ExpiresAt.Time,
		entities.InvitationResponse(row.Response),
		optionalTimeFromPg(row.RespondedAt),
		optionalTimeFromPg(row.ClaimedAt),
		row.CreatedAt.Time,
	)
}

// InvitationToCreateParams converts a domain Invitation into create
// params. New invitations are always created pending (response and
// responded_at are fixed in the SQL). claimed_at is mapped because a
// share-link self-add mints an invitation born claimed by its creator.
func InvitationToCreateParams(inv *entities.Invitation) generated.CreateInvitationParams {
	return generated.CreateInvitationParams{
		ID:        string(inv.ID()),
		MatchID:   string(inv.MatchID()),
		PlayerID:  string(inv.PlayerID()),
		TokenHash: inv.TokenHash(),
		ExpiresAt: pgtype.Timestamptz{Time: inv.ExpiresAt(), Valid: true},
		ClaimedAt: optionalTimeToPg(inv.ClaimedAt()),
		CreatedAt: pgtype.Timestamptz{Time: inv.CreatedAt(), Valid: true},
	}
}

// InvitationToUpdateResponseParams converts a domain Invitation into the
// params for persisting its current response.
func InvitationToUpdateResponseParams(inv *entities.Invitation) generated.UpdateInvitationResponseParams {
	return generated.UpdateInvitationResponseParams{
		ID:          string(inv.ID()),
		Response:    string(inv.Response()),
		RespondedAt: optionalTimeToPg(inv.RespondedAt()),
	}
}

// InvitationToClaimParams converts a claimed domain Invitation into the
// params for the conditional claim update (rotated token + claim time).
func InvitationToClaimParams(inv *entities.Invitation) generated.ClaimInvitationParams {
	return generated.ClaimInvitationParams{
		ID:        string(inv.ID()),
		TokenHash: inv.TokenHash(),
		ClaimedAt: optionalTimeToPg(inv.ClaimedAt()),
	}
}
