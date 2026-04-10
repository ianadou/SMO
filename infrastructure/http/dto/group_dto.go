package dto

import (
	"time"

	"github.com/ianadou/smo/domain/entities"
)

// CreateGroupRequest is the JSON body of POST /api/groups.
//
// Validation tags use the github.com/go-playground/validator/v10
// syntax that Gin supports out of the box. Format validation here
// (required, min, max) is duplicated with the domain entity's
// validation, but it lets the handler reject obviously invalid
// requests with a clear 400 before reaching the use case.
type CreateGroupRequest struct {
	Name        string `json:"name"         binding:"required,min=1,max=100"`
	OrganizerID string `json:"organizer_id" binding:"required"`
}

// GroupResponse is the JSON body returned for any group endpoint that
// returns a single group (POST /api/groups, GET /api/groups/:id).
type GroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OrganizerID string    `json:"organizer_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// GroupResponseFromEntity builds a GroupResponse from a domain Group
// entity. Use this in handlers right before serializing the response.
func GroupResponseFromEntity(group *entities.Group) GroupResponse {
	return GroupResponse{
		ID:          string(group.ID()),
		Name:        group.Name(),
		OrganizerID: string(group.OrganizerID()),
		CreatedAt:   group.CreatedAt(),
	}
}
