package dto

import (
	"time"

	"github.com/ianadou/smo/domain/entities"
)

// RegisterOrganizerRequest is the JSON body of POST /api/v1/auth/register.
type RegisterOrganizerRequest struct {
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
}

// LoginOrganizerRequest is the JSON body of POST /api/v1/auth/login.
type LoginOrganizerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// OrganizerResponse is the safe public projection of an Organizer:
// no password hash, no internal flags. Used as the response body of
// register and as the embedded `organizer` field of the login response.
type OrganizerResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrganizerResponseFromEntity converts a domain Organizer into the safe
// HTTP response form. Never call this with a Player or any other entity.
func OrganizerResponseFromEntity(o *entities.Organizer) OrganizerResponse {
	return OrganizerResponse{
		ID:          string(o.ID()),
		Email:       o.Email(),
		DisplayName: o.DisplayName(),
		CreatedAt:   o.CreatedAt(),
	}
}

// LoginResponse is the JSON body of POST /api/v1/auth/login on success:
// the bearer token plus a safe projection of the organizer.
type LoginResponse struct {
	Token     string            `json:"token"`
	Organizer OrganizerResponse `json:"organizer"`
}
