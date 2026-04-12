package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/player"
	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

// PlayerHandler exposes the Player aggregate over HTTP.
type PlayerHandler struct {
	createPlayer        *player.CreatePlayerUseCase
	getPlayer           *player.GetPlayerUseCase
	listPlayersByGroup  *player.ListPlayersByGroupUseCase
	updatePlayerRanking *player.UpdatePlayerRankingUseCase
}

// NewPlayerHandler builds a PlayerHandler.
func NewPlayerHandler(
	createPlayer *player.CreatePlayerUseCase,
	getPlayer *player.GetPlayerUseCase,
	listPlayersByGroup *player.ListPlayersByGroupUseCase,
	updatePlayerRanking *player.UpdatePlayerRankingUseCase,
) *PlayerHandler {
	return &PlayerHandler{
		createPlayer:        createPlayer,
		getPlayer:           getPlayer,
		listPlayersByGroup:  listPlayersByGroup,
		updatePlayerRanking: updatePlayerRanking,
	}
}

// Register wires the player routes.
func (h *PlayerHandler) Register(api *gin.RouterGroup) {
	players := api.Group("/players")
	players.POST("", h.Create)
	players.GET("/:id", h.Get)
	players.PATCH("/:id/ranking", h.UpdateRanking)

	api.GET("/groups/:id/players", h.ListByGroup)
}

// Create handles POST /api/players.
func (h *PlayerHandler) Create(c *gin.Context) {
	var req dto.CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	created, err := h.createPlayer.Execute(c.Request.Context(), player.CreatePlayerInput{
		GroupID: entities.GroupID(req.GroupID),
		Name:    req.Name,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusCreated, dto.PlayerResponseFromEntity(created))
}

// Get handles GET /api/players/:id.
func (h *PlayerHandler) Get(c *gin.Context) {
	id := entities.PlayerID(c.Param("id"))
	p, err := h.getPlayer.Execute(c.Request.Context(), id)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.PlayerResponseFromEntity(p))
}

// ListByGroup handles GET /api/groups/:id/players.
func (h *PlayerHandler) ListByGroup(c *gin.Context) {
	groupID := entities.GroupID(c.Param("id"))
	players, err := h.listPlayersByGroup.Execute(c.Request.Context(), groupID)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	responses := make([]dto.PlayerResponse, 0, len(players))
	for _, p := range players {
		responses = append(responses, dto.PlayerResponseFromEntity(p))
	}
	c.JSON(http.StatusOK, responses)
}

// UpdateRanking handles PATCH /api/players/:id/ranking.
func (h *PlayerHandler) UpdateRanking(c *gin.Context) {
	id := entities.PlayerID(c.Param("id"))

	var req dto.UpdatePlayerRankingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	updated, err := h.updatePlayerRanking.Execute(c.Request.Context(), id, req.Ranking)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.PlayerResponseFromEntity(updated))
}
