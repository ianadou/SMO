package handlers

import (
	"errors"
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

// Create handles POST /api/v1/groups. The organizer ID is taken from
// the JWT context, never from the request body — this prevents an
// authenticated organizer from creating a group on behalf of someone
// else (IDOR). The strict JSON decoder rejects bodies that include
// retired fields like `organizer_id`.
func (h *GroupHandler) Create(ctx *gin.Context) {
	var request dto.CreateGroupRequest
	if err := bindStrictJSON(ctx, &request); err != nil {
		if errors.Is(err, errUnknownField) {
			ctx.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	organizerID := middlewares.OrganizerIDFromContext(ctx.Request.Context())
	if organizerID == "" {
		ctx.JSON(http.StatusUnauthorized, httperrors.ErrorResponse{Error: "missing authenticated organizer"})
		return
	}

	createdGroup, err := h.createGroup.Execute(ctx.Request.Context(), group.CreateGroupInput{
		Name:        request.Name,
		OrganizerID: organizerID,
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
