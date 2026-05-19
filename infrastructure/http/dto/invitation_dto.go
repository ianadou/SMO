package dto

import (
	"time"

	"github.com/ianadou/smo/domain/entities"
)

// CreateInvitationRequest is the body of POST /api/invitations.
// ExpiresAt is optional; if zero the use case applies the default.
// PlayerID identifies the recipient of this invitation; the use case
// rejects players that do not belong to the match's group.
type CreateInvitationRequest struct {
	MatchID   string    `json:"match_id"             binding:"required"`
	PlayerID  string    `json:"player_id"            binding:"required"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// RespondInvitationRequest is the body of POST /api/invitations/respond.
// Token is the plain value shared by the organizer; Answer is the
// player's reply, restricted to "yes" or "no" (pending is the initial
// state, never something a client can set).
type RespondInvitationRequest struct {
	Token  string `json:"token"  binding:"required"`
	Answer string `json:"answer" binding:"required,oneof=yes no"`
}

// RespondInvitationResponse is the body returned by a successful
// POST /api/invitations/respond. It exposes only the mutable part of
// the invitation the player just changed.
type RespondInvitationResponse struct {
	Response    string     `json:"response"`
	RespondedAt *time.Time `json:"responded_at"`
}

// RespondInvitationResponseFromEntity projects the answered invitation
// into the respond response shape.
func RespondInvitationResponseFromEntity(inv *entities.Invitation) RespondInvitationResponse {
	return RespondInvitationResponse{
		Response:    string(inv.Response()),
		RespondedAt: inv.RespondedAt(),
	}
}

// ParticipantResponse is the JSON shape of one entry in the response of
// GET /matches/:id/participants. It mirrors entities.MatchParticipant.
type ParticipantResponse struct {
	PlayerID    string    `json:"player_id"`
	PlayerName  string    `json:"player_name"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

// ParticipantResponsesFromEntities converts a slice of MatchParticipant
// projections into the wire-format response.
func ParticipantResponsesFromEntities(participants []entities.MatchParticipant) []ParticipantResponse {
	out := make([]ParticipantResponse, 0, len(participants))
	for _, p := range participants {
		out = append(out, ParticipantResponse{
			PlayerID:    string(p.PlayerID),
			PlayerName:  p.PlayerName,
			ConfirmedAt: p.ConfirmedAt,
		})
	}
	return out
}

// InvitationResponse is the standard invitation representation.
// It does NOT include the plain token: once an invitation is created,
// the token is never returned again.
type InvitationResponse struct {
	ID          string     `json:"id"`
	MatchID     string     `json:"match_id"`
	PlayerID    string     `json:"player_id"`
	ExpiresAt   time.Time  `json:"expires_at"`
	Response    string     `json:"response"`
	RespondedAt *time.Time `json:"responded_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateInvitationResponse is the one-time response returned by
// POST /api/invitations. Unlike InvitationResponse, it includes the
// plain token. The plain token is shown ONCE at creation time.
type CreateInvitationResponse struct {
	InvitationResponse
	PlainToken string `json:"plain_token"`
}

// InvitationResponseFromEntity converts a domain Invitation into the
// standard response (no plain token).
func InvitationResponseFromEntity(inv *entities.Invitation) InvitationResponse {
	return InvitationResponse{
		ID:          string(inv.ID()),
		MatchID:     string(inv.MatchID()),
		PlayerID:    string(inv.PlayerID()),
		ExpiresAt:   inv.ExpiresAt(),
		Response:    string(inv.Response()),
		RespondedAt: inv.RespondedAt(),
		CreatedAt:   inv.CreatedAt(),
	}
}
