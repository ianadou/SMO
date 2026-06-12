package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/sharelink"
	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
	"github.com/ianadou/smo/infrastructure/http/middlewares"
)

// ShareLinkHandler exposes the match share link flow over HTTP: the
// organizer generates or revokes the single shareable link of a match,
// and anonymous visitors resolve it to claim a roster name or add
// themselves.
type ShareLinkHandler struct {
	generateShareLink   *sharelink.GenerateMatchShareLinkUseCase
	revokeShareLink     *sharelink.RevokeMatchShareLinkUseCase
	getShareLinkContext *sharelink.GetShareLinkContextUseCase
	claimInvitation     *sharelink.ClaimInvitationUseCase
	joinMatch           *sharelink.JoinMatchUseCase
}

// NewShareLinkHandler builds a ShareLinkHandler.
func NewShareLinkHandler(
	generateShareLink *sharelink.GenerateMatchShareLinkUseCase,
	revokeShareLink *sharelink.RevokeMatchShareLinkUseCase,
	getShareLinkContext *sharelink.GetShareLinkContextUseCase,
	claimInvitation *sharelink.ClaimInvitationUseCase,
	joinMatch *sharelink.JoinMatchUseCase,
) *ShareLinkHandler {
	return &ShareLinkHandler{
		generateShareLink:   generateShareLink,
		revokeShareLink:     revokeShareLink,
		getShareLinkContext: getShareLinkContext,
		claimInvitation:     claimInvitation,
		joinMatch:           joinMatch,
	}
}

// Register wires share link routes. Generation and revocation are
// organizer-only; resolving, claiming and joining are public — the
// share token in the URL is the whole credential.
func (h *ShareLinkHandler) Register(public, protected *gin.RouterGroup) {
	publicShare := public.Group("/share")
	publicShare.GET("/:token", h.Context)
	publicShare.POST("/:token/claim", h.Claim)
	publicShare.POST("/:token/join", h.Join)

	protected.POST("/matches/:id/share-link", h.Generate)
	protected.DELETE("/matches/:id/share-link", h.Revoke)
}

// Generate handles POST /api/v1/matches/:id/share-link. The response
// includes the one-time plain token; regenerating revokes the link
// already circulating.
func (h *ShareLinkHandler) Generate(c *gin.Context) {
	organizerID := middlewares.OrganizerIDFromContext(c.Request.Context())
	if organizerID == "" {
		c.JSON(http.StatusUnauthorized, httperrors.ErrorResponse{Error: "missing authenticated organizer"})
		return
	}

	result, err := h.generateShareLink.Execute(c.Request.Context(), entities.MatchID(c.Param("id")), organizerID)
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusCreated, dto.MatchShareLinkResponseFromResult(result))
}

// Revoke handles DELETE /api/v1/matches/:id/share-link.
func (h *ShareLinkHandler) Revoke(c *gin.Context) {
	organizerID := middlewares.OrganizerIDFromContext(c.Request.Context())
	if organizerID == "" {
		c.JSON(http.StatusUnauthorized, httperrors.ErrorResponse{Error: "missing authenticated organizer"})
		return
	}

	if err := h.revokeShareLink.Execute(c.Request.Context(), entities.MatchID(c.Param("id")), organizerID); err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.Status(http.StatusNoContent)
}

// Context handles GET /api/v1/share/:token. The only 404 is a dead
// link (unknown, revoked or expired — deliberately indistinguishable);
// a locked match is reported in the body through match_status.
func (h *ShareLinkHandler) Context(c *gin.Context) {
	pageContext, err := h.getShareLinkContext.Execute(c.Request.Context(), c.Param("token"))
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.ShareLinkContextResponseFromContext(pageContext))
}

// Claim handles POST /api/v1/share/:token/claim. On success the fresh
// personal invitation token is returned once.
func (h *ShareLinkHandler) Claim(c *gin.Context) {
	var req dto.ClaimShareInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	result, err := h.claimInvitation.Execute(c.Request.Context(), c.Param("token"), entities.PlayerID(req.PlayerID))
	if err != nil {
		status, message := mapShareActionError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.ClaimedInvitationResponse{InvitationToken: result.PlainToken})
}

// Join handles POST /api/v1/share/:token/join. On success the visitor's
// brand-new personal invitation token is returned once.
func (h *ShareLinkHandler) Join(c *gin.Context) {
	var req dto.JoinMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	result, err := h.joinMatch.Execute(c.Request.Context(), sharelink.JoinMatchInput{
		ShareToken: c.Param("token"),
		PlayerName: req.PlayerName,
	})
	if err != nil {
		status, message := mapShareActionError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusCreated, dto.ClaimedInvitationResponse{InvitationToken: result.PlainToken})
}

// mapShareActionError wraps the shared mapper for claim/join. A locked
// match is 423 here instead of the mapper's 409: the join page must
// tell "registrations are closed" apart from "someone beat you to this
// name", and both land on these two endpoints. The respond endpoint
// keeps its historical 409 for the same domain error.
func mapShareActionError(err error) (int, string) {
	if errors.Is(err, domainerrors.ErrInvitationLocked) {
		return http.StatusLocked, "match attendance is locked"
	}
	return httperrors.MapError(err)
}
