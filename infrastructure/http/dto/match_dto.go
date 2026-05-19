package dto

import (
	"time"

	"github.com/ianadou/smo/application/usecases/match"
	"github.com/ianadou/smo/domain/entities"
)

// CreateMatchRequest is the JSON body expected by POST /api/matches.
//
// ScheduledAt must be a valid RFC 3339 timestamp (e.g.,
// "2026-05-01T18:00:00Z"). The status is not in this payload because
// a new match always starts in "draft".
type CreateMatchRequest struct {
	GroupID     string    `json:"group_id"     binding:"required"`
	Title       string    `json:"title"        binding:"required"`
	Venue       string    `json:"venue"        binding:"required"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

// MatchResponse is the JSON body returned for any endpoint that
// describes a single match (Create, Get, transitions) and, in a list,
// for endpoints that return multiple matches.
type MatchResponse struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	Title       string    `json:"title"`
	Venue       string    `json:"venue"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"`
	ScoreA      *int      `json:"score_a"`
	ScoreB      *int      `json:"score_b"`
	CreatedAt   time.Time `json:"created_at"`
}

// MatchResponseFromEntity converts a domain Match into the HTTP
// response representation.
func MatchResponseFromEntity(m *entities.Match) MatchResponse {
	return MatchResponse{
		ID:          string(m.ID()),
		GroupID:     string(m.GroupID()),
		Title:       m.Title(),
		Venue:       m.Venue(),
		ScheduledAt: m.ScheduledAt(),
		Status:      string(m.Status()),
		ScoreA:      m.ScoreA(),
		ScoreB:      m.ScoreB(),
		CreatedAt:   m.CreatedAt(),
	}
}

// CompleteMatchRequest is the body of POST /api/v1/matches/:id/complete.
// Scores are pointers so a 0-0 draw is distinguishable from a missing
// field: "required" rejects nil (absent), "gte=0" rejects negatives.
type CompleteMatchRequest struct {
	ScoreA *int `json:"score_a" binding:"required,gte=0"`
	ScoreB *int `json:"score_b" binding:"required,gte=0"`
}

// GenerateTeamsRequest is the body of POST /api/v1/matches/:id/teams/generate.
type GenerateTeamsRequest struct {
	Strategy string `json:"strategy" binding:"required,oneof=random ranking"`
}

// SetTeamsRequest is the body of PUT /api/v1/matches/:id/teams. Both
// sides must be non-empty; the exact-partition and balance rules are
// enforced by the domain, not by binding tags.
type SetTeamsRequest struct {
	TeamA []string `json:"team_a" binding:"required,min=1"`
	TeamB []string `json:"team_b" binding:"required,min=1"`
}

// TeamMemberResponse is one player of a team in the GET /teams payload.
// PlayerName is empty on generate/set responses (those return ids only);
// GET resolves names via the JOIN read model.
type TeamMemberResponse struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Team       string `json:"team"`
	Slot       int    `json:"slot"`
}

// TeamMemberResponsesFromEntities maps the read-model projection to the
// HTTP shape. A nil slice yields an empty array, never null.
func TeamMemberResponsesFromEntities(members []entities.MatchTeamMember) []TeamMemberResponse {
	out := make([]TeamMemberResponse, 0, len(members))
	for _, m := range members {
		out = append(out, TeamMemberResponse{
			PlayerID:   string(m.PlayerID),
			PlayerName: m.PlayerName,
			Team:       m.Team,
			Slot:       m.Slot,
		})
	}
	return out
}

// FinalizeMatchResponse is the JSON body returned by POST
// /api/matches/:id/finalize. It includes the final match state, the
// elected MVP (if any), and the new ranking of every participant whose
// score changed during the calculation.
type FinalizeMatchResponse struct {
	Match           MatchResponse  `json:"match"`
	MVPPlayerID     *string        `json:"mvp_player_id"`
	UpdatedRankings map[string]int `json:"updated_rankings"`
}

// FinalizeMatchResponseFromOutput converts the use case output into
// the HTTP response representation.
func FinalizeMatchResponseFromOutput(out *match.FinalizeMatchOutput) FinalizeMatchResponse {
	var mvp *string
	if out.MVP != nil {
		s := string(*out.MVP)
		mvp = &s
	}

	rankings := make(map[string]int, len(out.UpdatedRankings))
	for playerID, ranking := range out.UpdatedRankings {
		rankings[string(playerID)] = ranking
	}

	return FinalizeMatchResponse{
		Match:           MatchResponseFromEntity(out.Match),
		MVPPlayerID:     mvp,
		UpdatedRankings: rankings,
	}
}
