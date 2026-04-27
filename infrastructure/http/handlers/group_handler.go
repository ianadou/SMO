package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/group"
	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

// GroupHandler exposes the Group HTTP endpoints. It depends on the
// use cases via concrete pointers; the use cases themselves depend
// only on the domain ports, so the architecture remains clean.
type GroupHandler struct {
	createGroup *group.CreateGroupUseCase
	getGroup    *group.GetGroupUseCase
}

// NewGroupHandler builds a GroupHandler with the given use cases.
func NewGroupHandler(
	createGroup *group.CreateGroupUseCase,
	getGroup *group.GetGroupUseCase,
) *GroupHandler {
	return &GroupHandler{
		createGroup: createGroup,
		getGroup:    getGroup,
	}
}

// Register attaches the group routes. Reads go on `public`, mutations
// on `protected` (which carries the JWTAuth middleware in production).
//
// Routes:
//
//	GET  /groups/:id   → Get      (public)
//	POST /groups       → Create   (protected — organizer only)
func (h *GroupHandler) Register(public, protected *gin.RouterGroup) {
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
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		ctx.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	ctx.JSON(http.StatusCreated, dto.GroupResponseFromEntity(createdGroup))
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
