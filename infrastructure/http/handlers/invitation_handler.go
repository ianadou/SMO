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
	acceptInvitation       *invitation.AcceptInvitationUseCase
}

// NewInvitationHandler builds an InvitationHandler.
func NewInvitationHandler(
	createInvitation *invitation.CreateInvitationUseCase,
	getInvitation *invitation.GetInvitationUseCase,
	listInvitationsByMatch *invitation.ListInvitationsByMatchUseCase,
	acceptInvitation *invitation.AcceptInvitationUseCase,
) *InvitationHandler {
	return &InvitationHandler{
		createInvitation:       createInvitation,
		getInvitation:          getInvitation,
		listInvitationsByMatch: listInvitationsByMatch,
		acceptInvitation:       acceptInvitation,
	}
}

// Register wires the invitation routes. Accept is under the invitations
// group with a dedicated sub-path to keep the URL predictable.
func (h *InvitationHandler) Register(api *gin.RouterGroup) {
	invitations := api.Group("/invitations")
	invitations.POST("", h.Create)
	invitations.POST("/accept", h.Accept)
	invitations.GET("/:id", h.Get)

	api.GET("/matches/:id/invitations", h.ListByMatch)
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

// Accept handles POST /api/invitations/accept.
func (h *InvitationHandler) Accept(c *gin.Context) {
	var req dto.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	inv, err := h.acceptInvitation.Execute(c.Request.Context(), req.Token)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}
	c.JSON(http.StatusOK, dto.InvitationResponseFromEntity(inv))
}
