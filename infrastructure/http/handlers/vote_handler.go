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
	getVoteContext   *vote.GetVoteContextUseCase
	getVote          *vote.GetVoteUseCase
	listVotesByMatch *vote.ListVotesByMatchUseCase
}

// NewVoteHandler builds the handler.
func NewVoteHandler(
	castVote *vote.CastVoteUseCase,
	getVoteContext *vote.GetVoteContextUseCase,
	getVote *vote.GetVoteUseCase,
	listVotesByMatch *vote.ListVotesByMatchUseCase,
) *VoteHandler {
	return &VoteHandler{
		castVote:         castVote,
		getVoteContext:   getVoteContext,
		getVote:          getVote,
		listVotesByMatch: listVotesByMatch,
	}
}

// Register wires vote routes. Cast and Context are public but
// token-authed: the invitation token in the body is the capability that
// both identifies and authenticates the voter. Raw vote reads are
// organizer-only: votes are anonymous between players (the vote page
// promises "vos coéquipiers ne sauront pas qui les a notés"), so
// players only ever see aggregates via the context endpoint.
func (h *VoteHandler) Register(public, protected *gin.RouterGroup) {
	votes := public.Group("/votes")
	votes.POST("", h.Cast)
	votes.POST("/context", h.Context)

	protected.GET("/votes/:id", h.Get)
	protected.GET("/matches/:id/votes", h.ListByMatch)
}

// Cast handles POST /api/v1/votes.
func (h *VoteHandler) Cast(c *gin.Context) {
	var req dto.CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}
	v, err := h.castVote.Execute(c.Request.Context(), vote.CastVoteInput{
		PlainToken: req.Token,
		VotedID:    entities.PlayerID(req.VotedID),
		Score:      req.Score,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusCreated, dto.VoteResponseFromEntity(v))
}

// Context handles POST /api/v1/votes/context. It resolves the bearer's
// token into the full vote-page view model.
func (h *VoteHandler) Context(c *gin.Context) {
	var req dto.VoteContextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}
	pageContext, err := h.getVoteContext.Execute(c.Request.Context(), req.Token)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.VoteContextResponseFromContext(pageContext))
}

// Get handles GET /api/v1/votes/:id (organizer only).
func (h *VoteHandler) Get(c *gin.Context) {
	v, err := h.getVote.Execute(c.Request.Context(), entities.VoteID(c.Param("id")))
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.VoteResponseFromEntity(v))
}

// ListByMatch handles GET /api/v1/matches/:id/votes (organizer only).
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
