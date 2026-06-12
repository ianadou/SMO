package entities

import (
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// MatchShareLinkID is the unique identifier of a MatchShareLink.
type MatchShareLinkID string

// MatchShareLink represents the single shareable link of a match: one
// URL the organizer drops in the group chat, through which any invitee
// claims their personal invitation (or adds themselves to the roster).
//
// Like invitations, the token is stored as a hash, never in clear: the
// plain token is only returned once at creation time. A link dies in
// two ways — it expires, or the organizer revokes it. Regenerating a
// link revokes the previous one, so a match has at most one active link.
type MatchShareLink struct {
	id        MatchShareLinkID
	matchID   MatchID
	tokenHash string
	expiresAt time.Time
	revokedAt *time.Time
	createdAt time.Time
}

// NewMatchShareLink builds a MatchShareLink after validating its inputs.
// It serves both creation (nil revokedAt) and rehydration from
// persistence (revokedAt as stored), like NewInvitation does.
//
// The tokenHash parameter must be a non-empty string produced by an
// infrastructure adapter (e.g., SHA-256). This entity does not validate
// the format of the hash, only that it is non-empty.
//
// The expiresAt parameter must be in the future relative to createdAt;
// otherwise the link would be invalid as soon as it is created.
func NewMatchShareLink(
	id MatchShareLinkID,
	matchID MatchID,
	tokenHash string,
	expiresAt time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) (*MatchShareLink, error) {
	if id == "" || matchID == "" || tokenHash == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	if expiresAt.IsZero() || !expiresAt.After(createdAt) {
		return nil, domainerrors.ErrInvalidDate
	}

	return &MatchShareLink{
		id:        id,
		matchID:   matchID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		revokedAt: revokedAt,
		createdAt: createdAt,
	}, nil
}

// ID returns the share link identifier.
func (l *MatchShareLink) ID() MatchShareLinkID { return l.id }

// MatchID returns the identifier of the match this link gives access to.
func (l *MatchShareLink) MatchID() MatchID { return l.matchID }

// TokenHash returns the hashed token. The plain token is never stored.
func (l *MatchShareLink) TokenHash() string { return l.tokenHash }

// ExpiresAt returns the expiration timestamp of the link.
func (l *MatchShareLink) ExpiresAt() time.Time { return l.expiresAt }

// RevokedAt returns the timestamp at which the link was revoked, or nil
// if it was never revoked.
func (l *MatchShareLink) RevokedAt() *time.Time { return l.revokedAt }

// CreatedAt returns the creation timestamp of the link.
func (l *MatchShareLink) CreatedAt() time.Time { return l.createdAt }

// IsActive reports whether the link can still be used at the given
// reference time: not revoked and not yet expired. Expiry is exclusive
// of the boundary, matching Invitation.IsExpired.
func (l *MatchShareLink) IsActive(now time.Time) bool {
	return l.revokedAt == nil && now.Before(l.expiresAt)
}

// Revoke kills the link so its token stops resolving.
//
// Only an active link can be revoked: a revoked or expired link is
// already dead, and the caller maps the refusal to "no active link for
// this match" without revealing which way it died.
func (l *MatchShareLink) Revoke(now time.Time) error {
	if !l.IsActive(now) {
		return domainerrors.ErrShareLinkInactive
	}

	l.revokedAt = &now
	return nil
}
