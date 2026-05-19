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
	generateTeams         *match.GenerateTeamsUseCase
	setTeams              *match.SetTeamsUseCase
	getMatchTeams         *match.GetMatchTeamsUseCase
}

// MatchHandlerDeps groups the use cases a MatchHandler needs, so the
// constructor takes one parameter instead of nine.
type MatchHandlerDeps struct {
	CreateMatch           *match.CreateMatchUseCase
	GetMatch              *match.GetMatchUseCase
	ListMatchesByGroup    *match.ListMatchesByGroupUseCase
	OpenMatch             *match.OpenMatchUseCase
	MarkTeamsReady        *match.MarkTeamsReadyUseCase
	StartMatch            *match.StartMatchUseCase
	CompleteMatch         *match.CompleteMatchUseCase
	FinalizeMatch         *match.FinalizeMatchUseCase
	ListMatchParticipants *invitation.ListMatchParticipantsUseCase
	GenerateTeams         *match.GenerateTeamsUseCase
	SetTeams              *match.SetTeamsUseCase
	GetMatchTeams         *match.GetMatchTeamsUseCase
}

// NewMatchHandler builds a MatchHandler with the full set of use cases.
func NewMatchHandler(deps MatchHandlerDeps) *MatchHandler {
	return &MatchHandler{
		createMatch:           deps.CreateMatch,
		getMatch:              deps.GetMatch,
		listMatchesByGroup:    deps.ListMatchesByGroup,
		openMatch:             deps.OpenMatch,
		markTeamsReady:        deps.MarkTeamsReady,
		startMatch:            deps.StartMatch,
		completeMatch:         deps.CompleteMatch,
		finalizeMatch:         deps.FinalizeMatch,
		listMatchParticipants: deps.ListMatchParticipants,
		generateTeams:         deps.GenerateTeams,
		setTeams:              deps.SetTeams,
		getMatchTeams:         deps.GetMatchTeams,
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
	protectedMatches.POST("/:id/teams/generate", h.GenerateTeams)
	protectedMatches.PUT("/:id/teams", h.SetTeams)
	protectedMatches.GET("/:id/teams", h.GetTeams)

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

// Complete handles POST /api/matches/:id/complete. Unlike the other
// transitions it carries a body: the final score, which is the result
// that closes the match and feeds the next assignment's winner rule.
func (h *MatchHandler) Complete(c *gin.Context) {
	var req dto.CompleteMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	m, err := h.completeMatch.Execute(c.Request.Context(),
		entities.MatchID(c.Param("id")), *req.ScoreA, *req.ScoreB)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.MatchResponseFromEntity(m))
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

// GenerateTeams handles POST /api/matches/:id/teams/generate. It builds
// the two teams from the confirmed participants using the requested
// strategy and returns the resulting composition (ids only).
func (h *MatchHandler) GenerateTeams(c *gin.Context) {
	var req dto.GenerateTeamsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	m, err := h.generateTeams.Execute(c.Request.Context(), entities.MatchID(c.Param("id")), req.Strategy)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.TeamMemberResponsesFromEntities(teamMembersOf(m)))
}

// SetTeams handles PUT /api/matches/:id/teams. It applies an explicit
// organizer-provided partition of the confirmed participants.
func (h *MatchHandler) SetTeams(c *gin.Context) {
	var req dto.SetTeamsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	m, err := h.setTeams.Execute(c.Request.Context(), entities.MatchID(c.Param("id")),
		toPlayerIDs(req.TeamA), toPlayerIDs(req.TeamB))
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.TeamMemberResponsesFromEntities(teamMembersOf(m)))
}

// GetTeams handles GET /api/matches/:id/teams. It returns the team
// membership with display names resolved via the JOIN read model.
func (h *MatchHandler) GetTeams(c *gin.Context) {
	members, err := h.getMatchTeams.Execute(c.Request.Context(), entities.MatchID(c.Param("id")))
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.TeamMemberResponsesFromEntities(members))
}

func toPlayerIDs(ids []string) []entities.PlayerID {
	out := make([]entities.PlayerID, len(ids))
	for i, id := range ids {
		out[i] = entities.PlayerID(id)
	}
	return out
}

// teamMembersOf flattens a match's two rosters into the read-model
// shape. Names are left empty: generate/set echo back ids only, and the
// frontend refetches GET /teams for display names. Slot is the index of
// the player within its team.
func teamMembersOf(m *entities.Match) []entities.MatchTeamMember {
	out := make([]entities.MatchTeamMember, 0, len(m.TeamA())+len(m.TeamB()))
	for i, id := range m.TeamA() {
		out = append(out, entities.MatchTeamMember{PlayerID: id, Team: "A", Slot: i})
	}
	for i, id := range m.TeamB() {
		out = append(out, entities.MatchTeamMember{PlayerID: id, Team: "B", Slot: i})
	}
	return out
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
