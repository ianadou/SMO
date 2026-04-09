package entities

import (
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// InvitationID is the unique identifier of an Invitation.
type InvitationID string

// Invitation represents an invitation token that allows a non-authenticated
// person to join a specific match (and later vote in it).
//
// The token is stored as a hash, never in clear: the plain token is only
// returned once at creation time and shown to the organizer who shares it
// with the invitee. When the invitee uses the token, the system hashes
// their input and looks up the matching invitation by hash.
//
// This entity holds the hash, the match it belongs to, the expiration date,
// and a usedAt timestamp that is set when the invitation is consumed.
// The token hashing itself happens in an infrastructure adapter, not here.
type Invitation struct {
	id        InvitationID
	matchID   MatchID
	tokenHash string
	expiresAt time.Time
	usedAt    *time.Time
	createdAt time.Time
}

// NewInvitation builds an Invitation after validating its inputs.
//
// The tokenHash parameter must be a non-empty string produced by an
// infrastructure adapter (e.g., bcrypt, SHA-256). This entity does not
// validate the format of the hash, only that it is non-empty.
//
// The expiresAt parameter must be in the future relative to createdAt;
// otherwise the invitation would be invalid as soon as it is created.
//
// The usedAt parameter is typically nil for new invitations and set to
// a non-nil value when the invitation is consumed.
func NewInvitation(
	id InvitationID,
	matchID MatchID,
	tokenHash string,
	expiresAt time.Time,
	usedAt *time.Time,
	createdAt time.Time,
) (*Invitation, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if matchID == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if tokenHash == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	if expiresAt.IsZero() || !expiresAt.After(createdAt) {
		return nil, domainerrors.ErrInvalidDate
	}

	return &Invitation{
		id:        id,
		matchID:   matchID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		usedAt:    usedAt,
		createdAt: createdAt,
	}, nil
}

// ID returns the invitation identifier.
func (i *Invitation) ID() InvitationID { return i.id }

// MatchID returns the identifier of the match this invitation grants access to.
func (i *Invitation) MatchID() MatchID { return i.matchID }

// TokenHash returns the hashed token. The plain token is never stored.
func (i *Invitation) TokenHash() string { return i.tokenHash }

// ExpiresAt returns the expiration timestamp of the invitation.
func (i *Invitation) ExpiresAt() time.Time { return i.expiresAt }

// UsedAt returns the timestamp at which the invitation was consumed,
// or nil if it has not been used yet.
func (i *Invitation) UsedAt() *time.Time { return i.usedAt }

// CreatedAt returns the creation timestamp of the invitation.
func (i *Invitation) CreatedAt() time.Time { return i.createdAt }

// IsUsed reports whether the invitation has been consumed.
func (i *Invitation) IsUsed() bool { return i.usedAt != nil }

// IsExpired reports whether the invitation has expired relative to the
// given reference time. Pass time.Now() in production code; tests can pass
// a fixed time to make assertions deterministic.
func (i *Invitation) IsExpired(now time.Time) bool {
	return !now.Before(i.expiresAt)
}
