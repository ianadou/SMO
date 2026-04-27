package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ianadou/smo/application/usecases/auth"
	"github.com/ianadou/smo/infrastructure/http/dto"
	httperrors "github.com/ianadou/smo/infrastructure/http/errors"
)

// AuthHandler exposes the authentication routes (register, login).
//
// Both routes are public: the JWTAuth middleware is NOT applied here,
// since you cannot present a JWT before login.
type AuthHandler struct {
	register *auth.RegisterOrganizerUseCase
	login    *auth.LoginOrganizerUseCase
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(
	register *auth.RegisterOrganizerUseCase,
	login *auth.LoginOrganizerUseCase,
) *AuthHandler {
	return &AuthHandler{register: register, login: login}
}

// Register attaches the auth routes under the given /api/v1 router group.
func (h *AuthHandler) Register(api *gin.RouterGroup) {
	authGroup := api.Group("/auth")
	authGroup.POST("/register", h.RegisterOrganizer)
	authGroup.POST("/login", h.Login)
}

// RegisterOrganizer handles POST /api/v1/auth/register.
func (h *AuthHandler) RegisterOrganizer(c *gin.Context) {
	var req dto.RegisterOrganizerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	organizer, err := h.register.Execute(c.Request.Context(), auth.RegisterOrganizerInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusCreated, dto.OrganizerResponseFromEntity(organizer))
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginOrganizerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.ErrorResponse{Error: "invalid request body"})
		return
	}

	out, err := h.login.Execute(c.Request.Context(), auth.LoginOrganizerInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		status, message := httperrors.MapError(err)
		c.JSON(status, httperrors.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:     out.Token,
		Organizer: dto.OrganizerResponseFromEntity(out.Organizer),
	})
}
