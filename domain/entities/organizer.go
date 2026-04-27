package entities

import (
	"net/mail"
	"strings"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

const (
	maxOrganizerDisplayNameLength = 50
)

// OrganizerID is declared in group.go (where it was first introduced
// as the FK on the Group aggregate). This file only adds the entity
// methods on top of that type.

// Organizer is the only authenticated actor of the SMO system. Organizers
// own groups and create matches. Players, by contrast, are token-only
// participants and have no Organizer counterpart.
//
// The password hash is held as an opaque string. The hashing scheme
// (bcrypt today) is owned by an infrastructure adapter behind the
// PasswordHasher port, so swapping algorithms does not require changing
// the entity.
type Organizer struct {
	id           OrganizerID
	email        string
	displayName  string
	passwordHash string
	createdAt    time.Time
}

// NewOrganizer builds an Organizer after validating its inputs.
//
// Email format is checked with net/mail.ParseAddress, which accepts the
// full RFC 5322 grammar — slightly more permissive than a regex but
// well-tested in the standard library.
func NewOrganizer(
	id OrganizerID,
	email string,
	displayName string,
	passwordHash string,
	createdAt time.Time,
) (*Organizer, error) {
	if id == "" {
		return nil, domainerrors.ErrInvalidID
	}

	trimmedEmail := strings.TrimSpace(email)
	if _, err := mail.ParseAddress(trimmedEmail); err != nil {
		return nil, domainerrors.ErrInvalidEmail
	}

	trimmedName := strings.TrimSpace(displayName)
	if trimmedName == "" || len(trimmedName) > maxOrganizerDisplayNameLength {
		return nil, domainerrors.ErrInvalidName
	}

	if passwordHash == "" {
		return nil, domainerrors.ErrInvalidPassword
	}

	if createdAt.IsZero() {
		return nil, domainerrors.ErrInvalidDate
	}

	return &Organizer{
		id:           id,
		email:        strings.ToLower(trimmedEmail),
		displayName:  trimmedName,
		passwordHash: passwordHash,
		createdAt:    createdAt,
	}, nil
}

// ID returns the organizer identifier.
func (o *Organizer) ID() OrganizerID { return o.id }

// Email returns the organizer email, normalized to lowercase.
func (o *Organizer) Email() string { return o.email }

// DisplayName returns the organizer display name.
func (o *Organizer) DisplayName() string { return o.displayName }

// PasswordHash returns the opaque hash of the organizer password.
// Callers must never log this value or expose it over the network.
func (o *Organizer) PasswordHash() string { return o.passwordHash }

// CreatedAt returns the organizer creation timestamp.
func (o *Organizer) CreatedAt() time.Time { return o.createdAt }
