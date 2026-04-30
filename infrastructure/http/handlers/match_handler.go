package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/invitation"
	"github.com/ianadou/smo/application/usecases/match"
	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

// MatchHandler exposes the Match aggregate over HTTP.
type MatchHandler struct {
	createMatch           *match.CreateMatchUseCase
	getMatch              *match.GetMatchUseCase
	listMatchesByGroup    *match.ListMatchesByGroupUseCase
	openMatch             *match.OpenMatchUseCase
	markTeamsReady        *match.MarkTeamsReadyUseCase
	startMatch            *match.StartMatchUseCase
	completeMatch         *match.CompleteMatchUseCase
	finalizeMatch         *match.FinalizeMatchUseCase
	listMatchParticipants *invitation.ListMatchParticipantsUseCase
}

// NewMatchHandler builds a MatchHandler with the full set of use cases.
func NewMatchHandler(
	createMatch *match.CreateMatchUseCase,
	getMatch *match.GetMatchUseCase,
	listMatchesByGroup *match.ListMatchesByGroupUseCase,
	openMatch *match.OpenMatchUseCase,
	markTeamsReady *match.MarkTeamsReadyUseCase,
	startMatch *match.StartMatchUseCase,
	completeMatch *match.CompleteMatchUseCase,
	finalizeMatch *match.FinalizeMatchUseCase,
	listMatchParticipants *invitation.ListMatchParticipantsUseCase,
) *MatchHandler {
	return &MatchHandler{
		createMatch:           createMatch,
		getMatch:              getMatch,
		listMatchesByGroup:    listMatchesByGroup,
		openMatch:             openMatch,
		markTeamsReady:        markTeamsReady,
		startMatch:            startMatch,
		completeMatch:         completeMatch,
		finalizeMatch:         finalizeMatch,
		listMatchParticipants: listMatchParticipants,
	}
}

// Register wires match routes. Reads go on `public`, mutations on
// `protected` (which carries JWTAuth in production).
func (h *MatchHandler) Register(public, protected *gin.RouterGroup) {
	publicMatches := public.Group("/matches")
	publicMatches.GET("/:id", h.Get)
	public.GET("/groups/:id/matches", h.ListByGroup)

	protectedMatches := protected.Group("/matches")
	protectedMatches.POST("", h.Create)
	protectedMatches.POST("/:id/open", h.Open)
	protectedMatches.POST("/:id/teams-ready", h.MarkTeamsReady)
	protectedMatches.POST("/:id/start", h.Start)
	protectedMatches.POST("/:id/complete", h.Complete)
	protectedMatches.POST("/:id/finalize", h.Finalize)

	// Participants of a match are NOT public: §3 Personas decided that an
	// invited player must not see other matches' rosters. Until the
	// dual-mode middleware lands (PR A.1 follow-up of #49 and ADR 0008),
	// this endpoint requires a JWT organizer; the cross-organizer leak
	// is a known temporary trade-off.
	protectedMatches.GET("/:id/participants", h.Participants)
}

// Create handles POST /api/matches.
func (h *MatchHandler) Create(c *gin.Context) {
	var req dto.CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	created, err := h.createMatch.Execute(c.Request.Context(), match.CreateMatchInput{
		GroupID:     entities.GroupID(req.GroupID),
		Title:       req.Title,
		Venue:       req.Venue,
		ScheduledAt: req.ScheduledAt,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusCreated, dto.MatchResponseFromEntity(created))
}

// Get handles GET /api/matches/:id.
func (h *MatchHandler) Get(c *gin.Context) {
	id := entities.MatchID(c.Param("id"))

	m, err := h.getMatch.Execute(c.Request.Context(), id)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.MatchResponseFromEntity(m))
}

// ListByGroup handles GET /api/groups/:id/matches.
func (h *MatchHandler) ListByGroup(c *gin.Context) {
	groupID := entities.GroupID(c.Param("id"))

	matches, err := h.listMatchesByGroup.Execute(c.Request.Context(), groupID)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	responses := make([]dto.MatchResponse, 0, len(matches))
	for _, m := range matches {
		responses = append(responses, dto.MatchResponseFromEntity(m))
	}
	c.JSON(http.StatusOK, responses)
}

// Open handles POST /api/matches/:id/open.
func (h *MatchHandler) Open(c *gin.Context) {
	h.runTransition(c, h.openMatch.Execute)
}

// MarkTeamsReady handles POST /api/matches/:id/teams-ready.
func (h *MatchHandler) MarkTeamsReady(c *gin.Context) {
	h.runTransition(c, h.markTeamsReady.Execute)
}

// Start handles POST /api/matches/:id/start.
func (h *MatchHandler) Start(c *gin.Context) {
	h.runTransition(c, h.startMatch.Execute)
}

// Complete handles POST /api/matches/:id/complete.
func (h *MatchHandler) Complete(c *gin.Context) {
	h.runTransition(c, h.completeMatch.Execute)
}

// Participants handles GET /api/matches/:id/participants. Returns the
// confirmed participants (used invitations) ordered by confirmation time.
func (h *MatchHandler) Participants(c *gin.Context) {
	matchID := entities.MatchID(c.Param("id"))
	participants, err := h.listMatchParticipants.Execute(c.Request.Context(), matchID)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.ParticipantResponsesFromEntities(participants))
}

// Finalize handles POST /api/matches/:id/finalize. Unlike the four other
// transitions, finalize returns a richer payload: the match itself plus
// the elected MVP and the per-player ranking deltas.
func (h *MatchHandler) Finalize(c *gin.Context) {
	id := entities.MatchID(c.Param("id"))

	out, err := h.finalizeMatch.Execute(c.Request.Context(), id)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.FinalizeMatchResponseFromOutput(out))
}

// runTransition is the shared body of all five transition handlers.
func (h *MatchHandler) runTransition(
	c *gin.Context,
	execute func(ctx context.Context, id entities.MatchID) (*entities.Match, error),
) {
	id := entities.MatchID(c.Param("id"))

	m, err := execute(c.Request.Context(), id)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.MatchResponseFromEntity(m))
}
