package token

import (
	"testing"
)

func TestService_GenerateToken_ProducesNonEmpty64HexString(t *testing.T) {
	t.Parallel()
	svc := New()

	token, err := svc.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(token))
	}
	for _, c := range token {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("non-hex char %q in token %q", c, token)
			break
		}
	}
}

func TestService_GenerateToken_ProducesUniqueTokens(t *testing.T) {
	t.Parallel()
	svc := New()
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		token, err := svc.GenerateToken()
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if seen[token] {
			t.Fatalf("duplicate token generated at iteration %d: %q", i, token)
		}
		seen[token] = true
	}
}

func TestService_HashToken_IsDeterministic(t *testing.T) {
	t.Parallel()
	svc := New()
	h1 := svc.HashToken("plain-token-xyz")
	h2 := svc.HashToken("plain-token-xyz")
	if h1 != h2 {
		t.Errorf("expected deterministic hash, got %q and %q", h1, h2)
	}
}

func TestService_HashToken_DifferentInputsProduceDifferentHashes(t *testing.T) {
	t.Parallel()
	svc := New()
	h1 := svc.HashToken("token-a")
	h2 := svc.HashToken("token-b")
	if h1 == h2 {
		t.Errorf("expected different hashes for different inputs, got %q for both", h1)
	}
}

func TestService_HashToken_Returns64HexChars(t *testing.T) {
	t.Parallel()
	svc := New()
	hash := svc.HashToken("any-input")
	if len(hash) != 64 {
		t.Errorf("expected 64 hex chars (SHA-256 hex), got %d", len(hash))
	}
}
