package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/vote"
	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

// VoteHandler exposes the Vote aggregate over HTTP.
type VoteHandler struct {
	castVote         *vote.CastVoteUseCase
	getVote          *vote.GetVoteUseCase
	listVotesByMatch *vote.ListVotesByMatchUseCase
}

// NewVoteHandler builds the handler.
func NewVoteHandler(
	castVote *vote.CastVoteUseCase,
	getVote *vote.GetVoteUseCase,
	listVotesByMatch *vote.ListVotesByMatchUseCase,
) *VoteHandler {
	return &VoteHandler{castVote: castVote, getVote: getVote, listVotesByMatch: listVotesByMatch}
}

// Register wires vote routes.
func (h *VoteHandler) Register(api *gin.RouterGroup) {
	votes := api.Group("/votes")
	votes.POST("", h.Cast)
	votes.GET("/:id", h.Get)

	api.GET("/matches/:id/votes", h.ListByMatch)
}

// Cast handles POST /api/votes.
func (h *VoteHandler) Cast(c *gin.Context) {
	var req dto.CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}
	v, err := h.castVote.Execute(c.Request.Context(), vote.CastVoteInput{
		MatchID: entities.MatchID(req.MatchID),
		VoterID: entities.PlayerID(req.VoterID),
		VotedID: entities.PlayerID(req.VotedID),
		Score:   req.Score,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusCreated, dto.VoteResponseFromEntity(v))
}

// Get handles GET /api/votes/:id.
func (h *VoteHandler) Get(c *gin.Context) {
	v, err := h.getVote.Execute(c.Request.Context(), entities.VoteID(c.Param("id")))
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.VoteResponseFromEntity(v))
}

// ListByMatch handles GET /api/matches/:id/votes.
func (h *VoteHandler) ListByMatch(c *gin.Context) {
	votes, err := h.listVotesByMatch.Execute(c.Request.Context(), entities.MatchID(c.Param("id")))
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	responses := make([]dto.VoteResponse, 0, len(votes))
	for _, v := range votes {
		responses = append(responses, dto.VoteResponseFromEntity(v))
	}
	c.JSON(http.StatusOK, responses)
}
