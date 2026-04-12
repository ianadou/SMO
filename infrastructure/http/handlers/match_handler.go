package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/match"
	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

// MatchHandler exposes the Match aggregate over HTTP.
type MatchHandler struct {
	createMatch        *match.CreateMatchUseCase
	getMatch           *match.GetMatchUseCase
	listMatchesByGroup *match.ListMatchesByGroupUseCase
	openMatch          *match.OpenMatchUseCase
	markTeamsReady     *match.MarkTeamsReadyUseCase
	startMatch         *match.StartMatchUseCase
	completeMatch      *match.CompleteMatchUseCase
	closeMatch         *match.CloseMatchUseCase
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
	closeMatch *match.CloseMatchUseCase,
) *MatchHandler {
	return &MatchHandler{
		createMatch:        createMatch,
		getMatch:           getMatch,
		listMatchesByGroup: listMatchesByGroup,
		openMatch:          openMatch,
		markTeamsReady:     markTeamsReady,
		startMatch:         startMatch,
		completeMatch:      completeMatch,
		closeMatch:         closeMatch,
	}
}

// Register wires match routes on the given API router group. The
// "list matches by group" route lives under /groups/:id/matches
// because it expresses a parent→child relation in REST terms.
func (h *MatchHandler) Register(api *gin.RouterGroup) {
	matches := api.Group("/matches")
	matches.POST("", h.Create)
	matches.GET("/:id", h.Get)

	matches.POST("/:id/open", h.Open)
	matches.POST("/:id/teams-ready", h.MarkTeamsReady)
	matches.POST("/:id/start", h.Start)
	matches.POST("/:id/complete", h.Complete)
	matches.POST("/:id/close", h.Close)

	api.GET("/groups/:id/matches", h.ListByGroup)
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

// Close handles POST /api/matches/:id/close.
func (h *MatchHandler) Close(c *gin.Context) {
	h.runTransition(c, h.closeMatch.Execute)
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
