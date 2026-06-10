package dto

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/ianadou/smo/application/usecases/invitation"
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

// InvitationContextRequest is the body of POST /api/v1/invitations/context.
// The token travels in the body (never the URL) because it is a bearer
// capability and URLs leak into logs, history and referrers.
type InvitationContextRequest struct {
	Token string `json:"token" binding:"required"`
}

// InvitationContextResponse is everything the player invitation page
// needs to render. ConfirmedInitials deliberately replaces participant
// names: the token bearer is unauthenticated, so confirmed teammates are
// summarised by initials only, never by name. State is the single field
// the frontend switches on; Response carries the invitee's own answer.
type InvitationContextResponse struct {
	OrganizerName     string    `json:"organizer_name"`
	GroupName         string    `json:"group_name"`
	MatchTitle        string    `json:"match_title"`
	Venue             string    `json:"venue"`
	ScheduledAt       time.Time `json:"scheduled_at"`
	MatchStatus       string    `json:"match_status"`
	Capacity          string    `json:"capacity"`
	ConfirmedCount    int       `json:"confirmed_count"`
	MaxParticipants   int       `json:"max_participants"`
	ConfirmedInitials []string  `json:"confirmed_initials"`
	Response          string    `json:"response"`
	ExpiresAt         time.Time `json:"expires_at"`
	State             string    `json:"state"`
}

// Invitation context states. Precedence is expired > locked >
// respondable: an expired invitation can never be answered even on a
// still-open match, and a locked match freezes answers regardless of
// expiry.
const (
	invitationStateExpired     = "expired"
	invitationStateLocked      = "locked"
	invitationStateRespondable = "respondable"
)

// InvitationContextResponseFromContext projects the assembled use case
// context into the wire shape, deriving the presentation-only fields
// (capacity label, confirmed initials, coarse state).
func InvitationContextResponseFromContext(c *invitation.PageContext) InvitationContextResponse {
	initials := make([]string, 0, len(c.ConfirmedNames))
	for _, name := range c.ConfirmedNames {
		initials = append(initials, deriveInitials(name))
	}

	return InvitationContextResponse{
		OrganizerName:     c.OrganizerName,
		GroupName:         c.GroupName,
		MatchTitle:        c.MatchTitle,
		Venue:             c.Venue,
		ScheduledAt:       c.ScheduledAt,
		MatchStatus:       string(c.MatchStatus),
		Capacity:          fmt.Sprintf("%d (%dv%d)", c.MaxParticipants, c.MaxParticipants/2, c.MaxParticipants/2),
		ConfirmedCount:    len(c.ConfirmedNames),
		MaxParticipants:   c.MaxParticipants,
		ConfirmedInitials: initials,
		Response:          string(c.Response),
		ExpiresAt:         c.ExpiresAt,
		State:             deriveInvitationState(c.Expired, c.Locked),
	}
}

func deriveInvitationState(expired, locked bool) string {
	switch {
	case expired:
		return invitationStateExpired
	case locked:
		return invitationStateLocked
	default:
		return invitationStateRespondable
	}
}

// deriveInitials returns up to two uppercased initials (first and last
// word) for a display name, so confirmed teammates can be shown as
// avatars without exposing their names.
func deriveInitials(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return ""
	}

	first := firstRuneUpper(words[0])
	if len(words) == 1 {
		return first
	}
	return first + firstRuneUpper(words[len(words)-1])
}

func firstRuneUpper(word string) string {
	for _, r := range word {
		return string(unicode.ToUpper(r))
	}
	return ""
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
