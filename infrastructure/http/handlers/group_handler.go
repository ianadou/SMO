package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/group"
	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
	"github.com/ianadou/smo/infrastructure/http/middlewares"
)

// GroupHandler exposes the Group HTTP endpoints. It depends on the
// use cases via concrete pointers; the use cases themselves depend
// only on the domain ports, so the architecture remains clean.
type GroupHandler struct {
	createGroup           *group.CreateGroupUseCase
	getGroup              *group.GetGroupUseCase
	listGroupsByOrganizer *group.ListGroupsByOrganizerUseCase
}

// NewGroupHandler builds a GroupHandler with the given use cases.
func NewGroupHandler(
	createGroup *group.CreateGroupUseCase,
	getGroup *group.GetGroupUseCase,
	listGroupsByOrganizer *group.ListGroupsByOrganizerUseCase,
) *GroupHandler {
	return &GroupHandler{
		createGroup:           createGroup,
		getGroup:              getGroup,
		listGroupsByOrganizer: listGroupsByOrganizer,
	}
}

// Register attaches the group routes. Reads go on `public`, mutations
// on `protected` (which carries the JWTAuth middleware in production).
//
// Routes:
//
//	GET  /groups       → List     (protected — organizer scoped via JWT)
//	GET  /groups/:id   → Get      (public)
//	POST /groups       → Create   (protected — organizer only)
func (h *GroupHandler) Register(public, protected *gin.RouterGroup) {
	protected.GET("/groups", h.List)
	public.GET("/groups/:id", h.Get)
	protected.POST("/groups", h.Create)
}

// Create handles POST /api/groups.
func (h *GroupHandler) Create(ctx *gin.Context) {
	var request dto.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	createdGroup, err := h.createGroup.Execute(ctx.Request.Context(), group.CreateGroupInput{
		Name:        request.Name,
		OrganizerID: entities.OrganizerID(request.OrganizerID),
		WebhookURL:  request.DiscordWebhookURL,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		ctx.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	ctx.JSON(http.StatusCreated, dto.GroupResponseFromEntity(createdGroup))
}

// List handles GET /api/groups. The result is scoped to the
// authenticated organizer (read from the JWT context, not from the
// query string), so the caller can never list someone else's groups.
func (h *GroupHandler) List(ctx *gin.Context) {
	organizerID := middlewares.OrganizerIDFromContext(ctx.Request.Context())

	groups, err := h.listGroupsByOrganizer.Execute(ctx.Request.Context(), organizerID)
	if err != nil {
		status, message := httperrors.MapError(err)
		ctx.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	response := make([]dto.GroupResponse, 0, len(groups))
	for _, g := range groups {
		response = append(response, dto.GroupResponseFromEntity(g))
	}
	ctx.JSON(http.StatusOK, response)
}

// Get handles GET /api/groups/:id.
func (h *GroupHandler) Get(ctx *gin.Context) {
	id := entities.GroupID(ctx.Param("id"))

	foundGroup, err := h.getGroup.Execute(ctx.Request.Context(), id)
	if err != nil {
		status, message := httperrors.MapError(err)
		ctx.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	ctx.JSON(http.StatusOK, dto.GroupResponseFromEntity(foundGroup))
}
