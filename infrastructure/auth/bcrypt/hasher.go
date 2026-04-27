// Package bcrypt provides a PasswordHasher implementation backed by
// golang.org/x/crypto/bcrypt.
package bcrypt

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

// Hasher is the production PasswordHasher used by the auth flow.
type Hasher struct {
	cost int
}

// New returns a Hasher with the given bcrypt cost. Production should
// use bcrypt.DefaultCost (10). Tests should use bcrypt.MinCost (4) to
// keep the suite fast.
func New(cost int) *Hasher {
	return &Hasher{cost: cost}
}

// Hash returns the bcrypt hash of the given plain password.
func (h *Hasher) Hash(plain string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hasher: hash: %w", err)
	}
	return string(out), nil
}

// Compare returns nil if the plain password matches the stored hash,
// ErrInvalidCredentials otherwise. Internal bcrypt errors are wrapped
// as ErrInvalidCredentials too: the caller (login flow) must not leak
// implementation detail back to the user.
func (h *Hasher) Compare(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return domainerrors.ErrInvalidCredentials
		}
		return fmt.Errorf("bcrypt hasher: compare: %w", err)
	}
	return nil
}
