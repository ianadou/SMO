package vote

// fakeTokenService hashes deterministically by prefixing, so tests can
// seed invitations whose stored hash matches a known plain token.
type fakeTokenService struct{}

func (fakeTokenService) GenerateToken() (string, error) {
	panic("not implemented in vote tests")
}

func (fakeTokenService) HashToken(plain string) string {
	return "hashed:" + plain
}
