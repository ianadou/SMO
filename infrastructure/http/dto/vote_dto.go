package dto

import (
	"time"

	"github.com/ianadou/smo/application/usecases/vote"
	"github.com/ianadou/smo/domain/entities"
)

// CastVoteRequest is the body of POST /api/v1/votes. The voter is
// derived server-side from the invitation token — the client never
// names itself, which is what makes vote spoofing impossible. The
// token travels in the body (never the URL) because it is a bearer
// capability and URLs leak into logs, history and referrers.
type CastVoteRequest struct {
	Token   string `json:"token"    binding:"required"`
	VotedID string `json:"voted_id" binding:"required"`
	Score   int    `json:"score"    binding:"required,min=1,max=5"`
}

// VoteContextRequest is the body of POST /api/v1/votes/context.
type VoteContextRequest struct {
	Token string `json:"token" binding:"required"`
}

// VoteContextResponse is everything the player vote page needs to
// render any of its lifecycle screens. Unlike the invitation context,
// teammate names ARE exposed here: the bearer is a verified confirmed
// participant rating people they played with, not an anonymous invitee.
// Raw votes never appear — only the bearer's own scores and, once the
// match is closed, per-player aggregates (votes stay anonymous).
type VoteContextResponse struct {
	GroupName   string                `json:"group_name"`
	MatchTitle  string                `json:"match_title"`
	Venue       string                `json:"venue"`
	ScheduledAt time.Time             `json:"scheduled_at"`
	Status      string                `json:"status"`
	ScoreA      *int                  `json:"score_a"`
	ScoreB      *int                  `json:"score_b"`
	Winner      *string               `json:"winner"`
	Voter       VoteContextVoter      `json:"voter"`
	Teammates   []VoteContextTeammate `json:"teammates"`
	VotersDone  int                   `json:"voters_done"`
	VotersTotal int                   `json:"voters_total"`
	Results     *VoteResultsResponse  `json:"results"`
}

// VoteContextVoter identifies the token bearer on the vote page.
type VoteContextVoter struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
	Team     string `json:"team"`
}

// VoteContextTeammate is one row of the rating list. YourScore carries
// the bearer's already-cast score (locks the row), nil when not rated.
type VoteContextTeammate struct {
	PlayerID        string `json:"player_id"`
	Name            string `json:"name"`
	Initials        string `json:"initials"`
	MatchesTogether int    `json:"matches_together"`
	YourScore       *int   `json:"your_score"`
}

// VoteResultsResponse carries the closed-match aggregates.
type VoteResultsResponse struct {
	Teammates []VoteTeammateResult `json:"teammates"`
	Self      VoteSelfResult       `json:"self"`
}

// VoteTeammateResult is one teammate's final aggregate line.
type VoteTeammateResult struct {
	PlayerID   string   `json:"player_id"`
	Name       string   `json:"name"`
	Initials   string   `json:"initials"`
	Average    float64  `json:"average"`
	VotesCount int      `json:"votes_count"`
	Delta      *float64 `json:"delta"`
}

// VoteSelfResult is the bearer's own aggregate. Average is null when
// nobody rated the bearer.
type VoteSelfResult struct {
	Average    *float64 `json:"average"`
	VotesCount int      `json:"votes_count"`
}

// VoteContextResponseFromContext projects the assembled use case
// context into the wire shape, deriving initials per player.
func VoteContextResponseFromContext(c *vote.PageContext) VoteContextResponse {
	teammates := make([]VoteContextTeammate, 0, len(c.Teammates))
	for _, teammate := range c.Teammates {
		teammates = append(teammates, VoteContextTeammate{
			PlayerID:        string(teammate.PlayerID),
			Name:            teammate.Name,
			Initials:        deriveInitials(teammate.Name),
			MatchesTogether: teammate.MatchesTogether,
			YourScore:       teammate.YourScore,
		})
	}

	response := VoteContextResponse{
		GroupName:   c.GroupName,
		MatchTitle:  c.MatchTitle,
		Venue:       c.Venue,
		ScheduledAt: c.ScheduledAt,
		Status:      string(c.Status),
		ScoreA:      c.ScoreA,
		ScoreB:      c.ScoreB,
		Winner:      teamSideToString(c.WinningSide),
		Voter: VoteContextVoter{
			PlayerID: string(c.Voter.PlayerID),
			Name:     c.Voter.Name,
			Initials: deriveInitials(c.Voter.Name),
			Team:     string(c.Voter.Team),
		},
		Teammates:   teammates,
		VotersDone:  c.VotersDone,
		VotersTotal: c.VotersTotal,
	}

	if c.Results != nil {
		results := &VoteResultsResponse{
			Teammates: make([]VoteTeammateResult, 0, len(c.Results.Teammates)),
			Self: VoteSelfResult{
				Average:    c.Results.Self.Average,
				VotesCount: c.Results.Self.VotesCount,
			},
		}
		for _, teammate := range c.Results.Teammates {
			results.Teammates = append(results.Teammates, VoteTeammateResult{
				PlayerID:   string(teammate.PlayerID),
				Name:       teammate.Name,
				Initials:   deriveInitials(teammate.Name),
				Average:    teammate.Average,
				VotesCount: teammate.VotesCount,
				Delta:      teammate.Delta,
			})
		}
		response.Results = results
	}

	return response
}

func teamSideToString(side *entities.TeamSide) *string {
	if side == nil {
		return nil
	}
	value := string(*side)
	return &value
}

// VoteResponse is the JSON representation of a Vote.
type VoteResponse struct {
	ID        string    `json:"id"`
	MatchID   string    `json:"match_id"`
	VoterID   string    `json:"voter_id"`
	VotedID   string    `json:"voted_id"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

// VoteResponseFromEntity converts a domain Vote into its HTTP response.
func VoteResponseFromEntity(v *entities.Vote) VoteResponse {
	return VoteResponse{
		ID:        string(v.ID()),
		MatchID:   string(v.MatchID()),
		VoterID:   string(v.VoterID()),
		VotedID:   string(v.VotedID()),
		Score:     v.Score(),
		CreatedAt: v.CreatedAt(),
	}
}
