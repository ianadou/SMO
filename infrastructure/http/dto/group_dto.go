package dto

import (
	"time"

	"github.com/ianadou/smo/domain/entities"
)

// CreateGroupRequest is the JSON body of POST /api/v1/groups.
//
// The organizer ID is intentionally NOT a field of this struct — it is
// derived from the authenticated JWT context by the handler. Accepting
// it from the body would let an authenticated organizer create groups
// owned by another organizer (IDOR). Clients that send `organizer_id`
// receive a 400 thanks to the strict JSON decoder rejecting unknown
// fields.
//
// Validation tags use the github.com/go-playground/validator/v10
// syntax that Gin supports out of the box. Format validation here
// (required, min, max) is duplicated with the domain entity's
// validation, but it lets the handler reject obviously invalid
// requests with a clear 400 before reaching the use case.
//
// DiscordWebhookURL is OPTIONAL and is the only field that carries a
// secret on the request side. The strict format checks live in the
// domain (entities.NewGroup → validateWebhookURL); the binding tag
// here is intentionally minimal to let the domain own the rules.
type CreateGroupRequest struct {
	Name              string `json:"name"                          binding:"required,min=1,max=100"`
	DiscordWebhookURL string `json:"discord_webhook_url,omitempty"`
}

// GroupResponse is the JSON body returned for any group endpoint that
// returns a single group (POST /api/groups, GET /api/groups/:id).
//
// The Discord webhook URL is NEVER returned in clear: it grants
// posting rights on the channel and must not leak through API
// responses, server logs, or browser extension caches. Clients only
// observe HasWebhook (boolean) — to update the webhook, an organizer
// re-submits the URL via a future PATCH endpoint.
type GroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OrganizerID string    `json:"organizer_id"`
	HasWebhook  bool      `json:"has_webhook"`
	CreatedAt   time.Time `json:"created_at"`
}

// GroupResponseFromEntity builds a GroupResponse from a domain Group
// entity. Use this in handlers right before serializing the response.
//
// HasWebhook is computed from the presence of a webhook URL on the
// entity; the URL itself is intentionally NOT copied over.
func GroupResponseFromEntity(group *entities.Group) GroupResponse {
	return GroupResponse{
		ID:          string(group.ID()),
		Name:        group.Name(),
		OrganizerID: string(group.OrganizerID()),
		HasWebhook:  group.WebhookURL() != "",
		CreatedAt:   group.CreatedAt(),
	}
}
