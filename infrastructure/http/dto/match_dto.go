package dto

import (
	"time"

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
	CreatedAt   time.Time `json:"created_at"`
}

// MatchResponseFromEntity converts a domain Match into the HTTP
// response representation.
func MatchResponseFromEntity(match *entities.Match) MatchResponse {
	return MatchResponse{
		ID:          string(match.ID()),
		GroupID:     string(match.GroupID()),
		Title:       match.Title(),
		Venue:       match.Venue(),
		ScheduledAt: match.ScheduledAt(),
		Status:      string(match.Status()),
		CreatedAt:   match.CreatedAt(),
	}
}
