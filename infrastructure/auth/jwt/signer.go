// Package jwt provides a JWTSigner implementation backed by
// github.com/golang-jwt/jwt/v5 with HMAC-SHA256.
package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// Signer issues and verifies HS256 JWTs that carry the organizer ID as
// the standard "sub" claim plus an expiration ("exp"). No custom claims
// are added: anything the server needs about the organizer is fetched
// from the database via the embedded ID.
type Signer struct {
	secret   []byte
	lifetime time.Duration
	now      func() time.Time
}

// New builds a Signer with the given secret and token lifetime.
//
// The secret must be cryptographically strong (≥ 32 bytes recommended)
// and is read from the JWT_SECRET environment variable in production.
// The lifetime is typically 24h for SMO; longer values weaken the
// session security, shorter values force more re-logins.
func New(secret string, lifetime time.Duration) *Signer {
	return &Signer{
		secret:   []byte(secret),
		lifetime: lifetime,
		now:      time.Now,
	}
}

// Sign issues a token for the given organizer. The token contains:
//   - sub: organizer ID
//   - iat: issued-at timestamp
//   - exp: expiration timestamp (now + lifetime)
func (s *Signer) Sign(organizerID entities.OrganizerID) (string, error) {
	now := s.now()
	claims := jwtv5.RegisteredClaims{
		Subject:   string(organizerID),
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(now.Add(s.lifetime)),
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("jwt signer: sign: %w", err)
	}
	return signed, nil
}

// Verify parses the given token and returns the embedded organizer ID.
// Any failure (malformed token, bad signature, wrong algorithm, expired,
// missing subject) is reported as ErrInvalidToken: the caller does not
// need to distinguish between failure modes.
func (s *Signer) Verify(token string) (entities.OrganizerID, error) {
	parsed, err := jwtv5.ParseWithClaims(token, &jwtv5.RegisteredClaims{}, s.keyFunc)
	if err != nil {
		return "", fmt.Errorf("jwt signer: verify: %w", domainerrors.ErrInvalidToken)
	}

	claims, ok := parsed.Claims.(*jwtv5.RegisteredClaims)
	if !ok || !parsed.Valid {
		return "", fmt.Errorf("jwt signer: verify: %w", domainerrors.ErrInvalidToken)
	}

	if claims.Subject == "" {
		return "", fmt.Errorf("jwt signer: missing subject: %w", domainerrors.ErrInvalidToken)
	}

	return entities.OrganizerID(claims.Subject), nil
}

// keyFunc enforces that incoming tokens use HS256, rejecting tokens
// signed with the "none" algorithm or any asymmetric algorithm a
// malicious caller might try to substitute.
func (s *Signer) keyFunc(token *jwtv5.Token) (any, error) {
	if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return s.secret, nil
}
