package dto

import "github.com/ianadou/smo/domain/entities"

// CreatePlayerRequest is the JSON body for POST /api/players.
// The ranking is intentionally not accepted here: new players always
// start at the default ranking.
type CreatePlayerRequest struct {
	GroupID string `json:"group_id" binding:"required"`
	Name    string `json:"name"     binding:"required"`
}

// UpdatePlayerRankingRequest is the JSON body for
// PATCH /api/players/:id/ranking.
type UpdatePlayerRankingRequest struct {
	Ranking int `json:"ranking" binding:"required"`
}

// PlayerResponse is the JSON representation of a Player.
type PlayerResponse struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
	Ranking int    `json:"ranking"`
}

// PlayerResponseFromEntity converts a domain Player into its HTTP
// response representation.
func PlayerResponseFromEntity(player *entities.Player) PlayerResponse {
	return PlayerResponse{
		ID:      string(player.ID()),
		GroupID: string(player.GroupID()),
		Name:    player.Name(),
		Ranking: player.Ranking(),
	}
}
