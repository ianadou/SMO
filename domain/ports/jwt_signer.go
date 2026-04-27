package ports

import (
	"github.com/ianadou/smo/domain/entities"
)

// JWTSigner abstracts the JSON Web Token signing scheme. The concrete
// implementation (HS256 via golang-jwt today) lives in
// infrastructure/auth/jwt.
//
// Sign issues a token for the given organizer. Verify parses a token
// and returns the embedded organizer ID, or ErrInvalidToken (wrapped)
// if the token is malformed, has an invalid signature, is expired, or
// carries unexpected claims.
type JWTSigner interface {
	Sign(organizerID entities.OrganizerID) (string, error)
	Verify(token string) (entities.OrganizerID, error)
}
