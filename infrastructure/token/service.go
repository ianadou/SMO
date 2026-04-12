// Package token provides the concrete implementation of the
// InvitationTokenService domain port.
//
// Tokens are generated as 32 random bytes encoded as hex (64 hex chars).
// They are hashed with SHA-256 and stored as the hex encoding of the
// digest (64 hex chars).
//
// SHA-256 is sufficient here because invitation tokens are already
// high-entropy random values (32 bytes = 256 bits of entropy). We do
// NOT need the slow hashing (bcrypt, argon2) used for user passwords,
// which exists to resist brute-force on low-entropy inputs. Fast hashing
// also keeps the hot path (invitation lookup by hash) fast.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// tokenByteLength is the number of random bytes generated per token.
// 32 bytes → 64 hex chars, 256 bits of entropy. Comfortably unguessable.
const tokenByteLength = 32

// Service is the production implementation of InvitationTokenService.
type Service struct{}

// New builds a Service. No configuration needed.
func New() *Service { return &Service{} }

// GenerateToken returns a fresh random 64-hex-char token.
func (s *Service) GenerateToken() (string, error) {
	buf := make([]byte, tokenByteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("token service: generate: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// HashToken returns the SHA-256 hash of the given plain token, encoded
// as a 64-hex-char string.
func (s *Service) HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
