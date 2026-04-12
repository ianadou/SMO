package ports

// InvitationTokenService produces and hashes invitation tokens.
//
// The plain token is what the organizer shares with the invitee; the
// hash is what the server persists. GenerateToken must use a
// cryptographically secure RNG to make tokens unguessable; HashToken
// must be deterministic so that the same plain input always produces
// the same hash for lookups by hash.
type InvitationTokenService interface {
	// GenerateToken returns a fresh random plain token. Callers must
	// treat the return value as a secret and never log or persist it.
	GenerateToken() (string, error)

	// HashToken returns the deterministic hash of the given plain token.
	// Used both when creating an invitation (hash the plain token
	// generated above) and when accepting one (hash the user-submitted
	// token to look it up by hash).
	HashToken(plain string) string
}
