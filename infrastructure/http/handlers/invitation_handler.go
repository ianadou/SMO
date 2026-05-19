package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/invitation"
	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

// InvitationHandler exposes the Invitation aggregate over HTTP.
type InvitationHandler struct {
	createInvitation       *invitation.CreateInvitationUseCase
	getInvitation          *invitation.GetInvitationUseCase
	listInvitationsByMatch *invitation.ListInvitationsByMatchUseCase
	respondToInvitation    *invitation.RespondToInvitationUseCase
}

// NewInvitationHandler builds an InvitationHandler.
func NewInvitationHandler(
	createInvitation *invitation.CreateInvitationUseCase,
	getInvitation *invitation.GetInvitationUseCase,
	listInvitationsByMatch *invitation.ListInvitationsByMatchUseCase,
	respondToInvitation *invitation.RespondToInvitationUseCase,
) *InvitationHandler {
	return &InvitationHandler{
		createInvitation:       createInvitation,
		getInvitation:          getInvitation,
		listInvitationsByMatch: listInvitationsByMatch,
		respondToInvitation:    respondToInvitation,
	}
}

// Register wires invitation routes. Token-authed actions (Respond) and
// public reads of a single invitation go on `public`. Organizer-only
// operations (Create, ListByMatch) go on `protected`.
func (h *InvitationHandler) Register(public, protected *gin.RouterGroup) {
	publicInvitations := public.Group("/invitations")
	publicInvitations.POST("/respond", h.Respond)
	publicInvitations.GET("/:id", h.Get)

	protected.POST("/invitations", h.Create)
	protected.GET("/matches/:id/invitations", h.ListByMatch)
}

// Create handles POST /api/invitations.
//
// The response includes the one-time plain token. Clients must surface
// it to the end user immediately; subsequent reads of this invitation
// will only return the hash.
func (h *InvitationHandler) Create(c *gin.Context) {
	var req dto.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	result, err := h.createInvitation.Execute(c.Request.Context(), invitation.CreateInvitationInput{
		MatchID:   entities.MatchID(req.MatchID),
		PlayerID:  entities.PlayerID(req.PlayerID),
		ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	response := dto.CreateInvitationResponse{
		InvitationResponse: dto.InvitationResponseFromEntity(result.Invitation),
		PlainToken:         result.PlainToken,
	}
	c.JSON(http.StatusCreated, response)
}

// Get handles GET /api/invitations/:id.
func (h *InvitationHandler) Get(c *gin.Context) {
	id := entities.InvitationID(c.Param("id"))
	inv, err := h.getInvitation.Execute(c.Request.Context(), id)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.InvitationResponseFromEntity(inv))
}

// ListByMatch handles GET /api/matches/:id/invitations.
func (h *InvitationHandler) ListByMatch(c *gin.Context) {
	matchID := entities.MatchID(c.Param("id"))
	invitations, err := h.listInvitationsByMatch.Execute(c.Request.Context(), matchID)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	responses := make([]dto.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		responses = append(responses, dto.InvitationResponseFromEntity(inv))
	}
	c.JSON(http.StatusOK, responses)
}

// Respond handles POST /api/invitations/respond. The player submits
// their plain token and an answer ("yes" or "no"); the answer can be
// changed until the match locks attendance.
func (h *InvitationHandler) Respond(c *gin.Context) {
	var req dto.RespondInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	inv, err := h.respondToInvitation.Execute(
		c.Request.Context(),
		req.Token,
		entities.InvitationResponse(req.Answer),
	)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.RespondInvitationResponseFromEntity(inv))
}
