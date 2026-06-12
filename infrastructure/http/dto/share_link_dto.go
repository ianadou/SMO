package dto

import (
	"fmt"
	"time"

	"github.com/ianadou/smo/application/usecases/sharelink"
)

// MatchShareLinkResponse is the one-time response returned by
// POST /api/v1/matches/:id/share-link. Token is the plain share token,
// shown ONCE at generation time; subsequent reads only ever see the
// hash. The full URL is built by the frontend from its own origin, so
// the backend never has to know where the SPA is served from.
type MatchShareLinkResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MatchShareLinkResponseFromResult projects the generation result into
// the wire shape.
func MatchShareLinkResponseFromResult(result *sharelink.GenerateMatchShareLinkResult) MatchShareLinkResponse {
	return MatchShareLinkResponse{
		Token:     result.PlainToken,
		ExpiresAt: result.ShareLink.ExpiresAt(),
	}
}

// ClaimShareInvitationRequest is the body of POST /api/v1/share/:token/claim.
type ClaimShareInvitationRequest struct {
	PlayerID string `json:"player_id" binding:"required"`
}

// JoinMatchRequest is the body of POST /api/v1/share/:token/join.
type JoinMatchRequest struct {
	PlayerName string `json:"player_name" binding:"required"`
}

// ClaimedInvitationResponse is the body returned by a successful claim
// or join: the freshly minted personal invitation token, shown ONCE.
type ClaimedInvitationResponse struct {
	InvitationToken string `json:"invitation_token"`
}

// ShareRosterEntryResponse is one selectable name on the share page.
// State is claimable, claimed or responded; the page locks the row for
// the last two.
type ShareRosterEntryResponse struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	State      string `json:"state"`
}

// ShareLinkContextResponse is everything the public join page needs to
// render. Like the invitation context, confirmed participants are
// summarised by initials only: the visitor is unauthenticated. The
// roster names are exposed on purpose — picking your own name IS the
// feature. MatchID lets the page key its local token stash per match.
type ShareLinkContextResponse struct {
	MatchID           string                     `json:"match_id"`
	OrganizerName     string                     `json:"organizer_name"`
	GroupName         string                     `json:"group_name"`
	MatchTitle        string                     `json:"match_title"`
	Venue             string                     `json:"venue"`
	ScheduledAt       time.Time                  `json:"scheduled_at"`
	MatchStatus       string                     `json:"match_status"`
	Capacity          string                     `json:"capacity"`
	ConfirmedCount    int                        `json:"confirmed_count"`
	MaxParticipants   int                        `json:"max_participants"`
	ConfirmedInitials []string                   `json:"confirmed_initials"`
	Roster            []ShareRosterEntryResponse `json:"roster"`
}

// ShareLinkContextResponseFromContext projects the assembled use case
// context into the wire shape, deriving the presentation-only fields
// the same way the invitation context does.
func ShareLinkContextResponseFromContext(c *sharelink.PageContext) ShareLinkContextResponse {
	initials := make([]string, 0, len(c.ConfirmedNames))
	for _, name := range c.ConfirmedNames {
		initials = append(initials, deriveInitials(name))
	}

	roster := make([]ShareRosterEntryResponse, 0, len(c.Roster))
	for _, entry := range c.Roster {
		roster = append(roster, ShareRosterEntryResponse{
			PlayerID:   string(entry.PlayerID),
			PlayerName: entry.PlayerName,
			State:      string(entry.State),
		})
	}

	return ShareLinkContextResponse{
		MatchID:           string(c.MatchID),
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
		Roster:            roster,
	}
}
