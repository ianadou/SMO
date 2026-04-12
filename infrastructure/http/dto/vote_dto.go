package dto

import (
	"time"

	"github.com/ianadou/smo/domain/entities"
)

// CastVoteRequest is the body of POST /api/votes.
type CastVoteRequest struct {
	MatchID string `json:"match_id"  binding:"required"`
	VoterID string `json:"voter_id"  binding:"required"`
	VotedID string `json:"voted_id"  binding:"required"`
	Score   int    `json:"score"     binding:"required,min=1,max=5"`
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
