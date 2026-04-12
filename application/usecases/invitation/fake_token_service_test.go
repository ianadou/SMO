package invitation

import (
	"crypto/sha256"
	"encoding/hex"
)

// fakeTokenService is a deterministic test double: GenerateToken returns
// the next pre-configured plain token, HashToken is real SHA-256 to keep
// the hash consistency property that the production code relies on.
type fakeTokenService struct {
	tokens []string
	index  int
}

func newFakeTokenService(tokens ...string) *fakeTokenService {
	return &fakeTokenService{tokens: tokens}
}

func (s *fakeTokenService) GenerateToken() (string, error) {
	if s.index >= len(s.tokens) {
		panic("fakeTokenService: ran out of tokens")
	}
	token := s.tokens[s.index]
	s.index++
	return token, nil
}

func (s *fakeTokenService) HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
