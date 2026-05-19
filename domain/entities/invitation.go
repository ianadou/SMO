package entities

import (
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// InvitationID is the unique identifier of an Invitation.
type InvitationID string

// InvitationResponse is the player's answer to an invitation. It starts
// at "pending" and can be set (and changed) to "yes" or "no" until the
// match locks attendance.
type InvitationResponse string

const (
	// InvitationResponsePending is the initial state: the player has not
	// answered yet. It is never a settable answer.
	InvitationResponsePending InvitationResponse = "pending"

	// InvitationResponseYes means the player confirmed attendance.
	InvitationResponseYes InvitationResponse = "yes"

	// InvitationResponseNo means the player declined.
	InvitationResponseNo InvitationResponse = "no"
)

// Invitation represents an invitation token that allows a non-authenticated
// person to join a specific match (and later vote in it).
//
// The token is stored as a hash, never in clear: the plain token is only
// returned once at creation time and shown to the organizer who shares it
// with the invitee. When the invitee uses the token, the system hashes
// their input and looks up the matching invitation by hash.
//
// An invitation carries a mutable response: the invitee can answer yes or
// no, and change their mind, until the match locks attendance. The
// response invariant is strict: respondedAt is non-nil if and only if
// the response has been settled to yes or no; a pending invitation always
// has a nil respondedAt.
type Invitation struct {
	id          InvitationID
	matchID     MatchID
	playerID    PlayerID
	tokenHash   string
	expiresAt   time.Time
	response    InvitationResponse
	respondedAt *time.Time
	createdAt   time.Time
}

// NewInvitation builds an Invitation after validating its inputs.
//
// The tokenHash parameter must be a non-empty string produced by an
// infrastructure adapter (e.g., SHA-256). This entity does not validate
// the format of the hash, only that it is non-empty.
//
// The expiresAt parameter must be in the future relative to createdAt;
// otherwise the invitation would be invalid as soon as it is created.
//
// The response must be one of InvitationResponsePending/Yes/No. The
// respondedAt invariant is enforced: pending requires a nil respondedAt,
// while yes/no require a non-nil respondedAt (the moment the answer was
// recorded).
func NewInvitation(
	id InvitationID,
	matchID MatchID,
	playerID PlayerID,
	tokenHash string,
	expiresAt time.Time,
	response InvitationResponse,
	respondedAt *time.Time,
	createdAt time.Time,
) (*Invitation, error) {
	if id == "" || matchID == "" || playerID == "" || tokenHash == "" {
		return nil, domainerrors.ErrInvalidID
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	if expiresAt.IsZero() || !expiresAt.After(createdAt) {
		return nil, domainerrors.ErrInvalidDate
	}

	if err := validateResponseInvariant(response, respondedAt); err != nil {
		return nil, err
	}

	return &Invitation{
		id:          id,
		matchID:     matchID,
		playerID:    playerID,
		tokenHash:   tokenHash,
		expiresAt:   expiresAt,
		response:    response,
		respondedAt: respondedAt,
		createdAt:   createdAt,
	}, nil
}

// validateResponseInvariant enforces that the response is a known value
// and that respondedAt is set if and only if the response is settled.
func validateResponseInvariant(response InvitationResponse, respondedAt *time.Time) error {
	switch response {
	case InvitationResponsePending:
		if respondedAt != nil {
			return domainerrors.ErrInvalidInvitationResponse
		}
	case InvitationResponseYes, InvitationResponseNo:
		if respondedAt == nil {
			return domainerrors.ErrInvalidInvitationResponse
		}
	default:
		return domainerrors.ErrInvalidInvitationResponse
	}
	return nil
}

// ID returns the invitation identifier.
func (i *Invitation) ID() InvitationID { return i.id }

// MatchID returns the identifier of the match this invitation grants access to.
func (i *Invitation) MatchID() MatchID { return i.matchID }

// PlayerID returns the identifier of the player this invitation was issued for.
// Each invitation is one-to-one with a player; the organizer picks the recipient
// at creation time.
func (i *Invitation) PlayerID() PlayerID { return i.playerID }

// TokenHash returns the hashed token. The plain token is never stored.
func (i *Invitation) TokenHash() string { return i.tokenHash }

// ExpiresAt returns the expiration timestamp of the invitation.
func (i *Invitation) ExpiresAt() time.Time { return i.expiresAt }

// Response returns the current answer (pending, yes, or no).
func (i *Invitation) Response() InvitationResponse { return i.response }

// RespondedAt returns the timestamp at which the invitation was answered,
// or nil if it is still pending.
func (i *Invitation) RespondedAt() *time.Time { return i.respondedAt }

// CreatedAt returns the creation timestamp of the invitation.
func (i *Invitation) CreatedAt() time.Time { return i.createdAt }

// IsConfirmed reports whether the player confirmed attendance (response
// is "yes"). This is the projection that drives match participation.
func (i *Invitation) IsConfirmed() bool { return i.response == InvitationResponseYes }

// IsExpired reports whether the invitation has expired relative to the
// given reference time. Pass time.Now() in production code; tests can pass
// a fixed time to make assertions deterministic.
func (i *Invitation) IsExpired(now time.Time) bool {
	return !now.Before(i.expiresAt)
}

// Respond records (or changes) the player's answer.
//
// Only "yes" or "no" are valid answers: "pending" is the initial state,
// not something a caller can set. The transition is rejected when the
// invitation has expired, or when the match has locked attendance (the
// caller decides "locked" from the match status and passes it in).
//
// Responding with the same answer again is allowed and refreshes
// respondedAt — the operation is idempotent on the response value.
//
// Expiration takes priority over the lock: an expired invitation is the
// more actionable signal for the player even if the match also locked.
func (i *Invitation) Respond(answer InvitationResponse, now time.Time, locked bool) error {
	if answer != InvitationResponseYes && answer != InvitationResponseNo {
		return domainerrors.ErrInvalidInvitationResponse
	}
	if i.IsExpired(now) {
		return domainerrors.ErrInvitationExpired
	}
	if locked {
		return domainerrors.ErrInvitationLocked
	}

	i.response = answer
	i.respondedAt = &now
	return nil
}
