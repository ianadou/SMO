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

// AcceptInvitationRequest is the body of POST /api/invitations/accept.
// The token is the plain value shared by the organizer.
type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}

// InvitationResponse is the standard invitation representation.
// It does NOT include the plain token: once an invitation is created,
// the token is never returned again.
type InvitationResponse struct {
	ID        string     `json:"id"`
	MatchID   string     `json:"match_id"`
	PlayerID  string     `json:"player_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
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
		ID:        string(inv.ID()),
		MatchID:   string(inv.MatchID()),
		PlayerID:  string(inv.PlayerID()),
		ExpiresAt: inv.ExpiresAt(),
		UsedAt:    inv.UsedAt(),
		CreatedAt: inv.CreatedAt(),
	}
}
