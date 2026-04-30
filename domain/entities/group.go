package entities

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const (
	maxGroupNameLength  = 100
	maxWebhookURLLength = 2048
	discordWebhookHost  = "discord.com"
)

// discordWebhookPathRe matches the canonical Discord webhook URL path
// shape: /api/webhooks/<numeric_id>/<token>. Tokens are alphanumeric
// plus dash and underscore in current Discord docs; tightening this
// regex is preferable to a permissive `[^/]+` since the surface is a
// security boundary for outbound notifications.
var discordWebhookPathRe = regexp.MustCompile(`^/api/webhooks/\d+/[\w-]+$`)

// GroupID is the unique identifier of a Group.
type GroupID string

// OrganizerID is the unique identifier of an Organizer.
// Defined here because Group references it; the Organizer entity is
// declared in its own file.
type OrganizerID string

// Group represents a collection of players that play matches together.
// A group is owned by exactly one Organizer.
//
// webhookURL is the Discord channel webhook for "teams ready" match
// notifications. Empty when the group has no Discord channel
// configured. Treated as a secret: never returned in HTTP responses
// in clear, never written to logs.
type Group struct {
	id          GroupID
	name        string
	organizerID OrganizerID
	webhookURL  string
	createdAt   time.Time
}

// NewGroup builds a Group after validating its inputs. webhookURL may
// be empty (group without Discord), but if non-empty it must satisfy
// the rules in validateWebhookURL.
func NewGroup(
	id GroupID,
	name string,
	organizerID OrganizerID,
	webhookURL string,
	createdAt time.Time,
) (*Group, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > maxGroupNameLength {
		return nil, domainerrors.ErrInvalidName
	}

	if organizerID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if err := validateWebhookURL(webhookURL); err != nil {
		return nil, err
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	return &Group{
		id:          id,
		name:        trimmedName,
		organizerID: organizerID,
		webhookURL:  webhookURL,
		createdAt:   createdAt,
	}, nil
}

// validateWebhookURL enforces the strict rules a Discord webhook URL
// must satisfy when non-empty:
//
//  1. length ≤ maxWebhookURLLength characters
//  2. no ASCII control characters (CR/LF/NUL/etc. — header injection)
//  3. parsable by url.Parse
//  4. scheme is exactly "https" (no http, no other)
//  5. no embedded credentials (no userinfo)
//  6. host is exactly discord.com (rejects discordapp.com legacy and
//     any malicious lookalike)
//  7. path matches the canonical /api/webhooks/<id>/<token> shape
//
// Empty input is accepted: a group is allowed to have no Discord
// webhook configured (notifications are opt-in).
func validateWebhookURL(s string) error {
	if s == "" {
		return nil
	}
	if len(s) > maxWebhookURLLength {
		return domainerrors.ErrInvalidWebhookURL
	}
	if containsControlChar(s) {
		return domainerrors.ErrInvalidWebhookURL
	}
	parsed, err := url.Parse(s)
	if err != nil {
		return domainerrors.ErrInvalidWebhookURL
	}
	if parsed.Scheme != "https" {
		return domainerrors.ErrInvalidWebhookURL
	}
	if parsed.User != nil {
		return domainerrors.ErrInvalidWebhookURL
	}
	if parsed.Host != discordWebhookHost {
		return domainerrors.ErrInvalidWebhookURL
	}
	if !discordWebhookPathRe.MatchString(parsed.Path) {
		return domainerrors.ErrInvalidWebhookURL
	}
	return nil
}

// containsControlChar reports whether s contains any ASCII control
// character (0x00-0x1F or 0x7F). Catches CR, LF, NUL and friends used
// in header-injection payloads disguised as URLs.
func containsControlChar(s string) bool {
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

// ID returns the group identifier.
func (g *Group) ID() GroupID { return g.id }

// Name returns the group name.
func (g *Group) Name() string { return g.name }

// OrganizerID returns the identifier of the organizer who owns this group.
func (g *Group) OrganizerID() OrganizerID { return g.organizerID }

// WebhookURL returns the Discord webhook URL configured for this group,
// or empty string if the group has no Discord channel.
//
// Callers must treat the returned value as a secret: never log it,
// never include it in HTTP responses in clear. The HTTP DTO layer is
// responsible for masking (exposing only `has_webhook bool`).
func (g *Group) WebhookURL() string { return g.webhookURL }

// CreatedAt returns the creation timestamp of the group.
func (g *Group) CreatedAt() time.Time { return g.createdAt }
