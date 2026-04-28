package bcrypt_test

import (
	"errors"
	"testing"

	xbcrypt "golang.org/x/crypto/bcrypt"

	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/auth/bcrypt"
)

// Use bcrypt.MinCost (4) in tests to keep the suite fast: every Hash
// call is otherwise ~100ms at the default cost.

func TestHasher_HashAndCompare_RoundTrip(t *testing.T) {
	t.Parallel()
	hasher := bcrypt.New(xbcrypt.MinCost)

	hash, err := hasher.Hash("super-secret-password")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if hash == "super-secret-password" {
		t.Errorf("expected hash to differ from plain password")
	}

	if compareErr := hasher.Compare(hash, "super-secret-password"); compareErr != nil {
		t.Errorf("expected matching password to compare ok, got %v", compareErr)
	}
}

func TestHasher_Compare_ReturnsErrInvalidCredentials_WhenPasswordIsWrong(t *testing.T) {
	t.Parallel()
	hasher := bcrypt.New(xbcrypt.MinCost)

	hash, _ := hasher.Hash("correct")
	err := hasher.Compare(hash, "wrong")

	if !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestHasher_Hash_ReturnsError_WhenCostIsOutOfRange(t *testing.T) {
	t.Parallel()
	// bcrypt.GenerateFromPassword rejects costs above MaxCost (31).
	// Hasher must wrap that error rather than panic or silently
	// accept a degraded password hash.
	hasher := bcrypt.New(xbcrypt.MaxCost + 1)

	hash, err := hasher.Hash("anything")

	if err == nil {
		t.Fatalf("expected error for out-of-range cost, got nil")
	}
	if hash != "" {
		t.Errorf("expected empty hash on error, got %q", hash)
	}
}

func TestHasher_Compare_ReturnsWrappedError_OnMalformedHash(t *testing.T) {
	t.Parallel()
	// A non-bcrypt string is neither a valid hash nor "wrong password".
	// bcrypt returns a parse error; the hasher must surface it (NOT
	// wrap it as ErrInvalidCredentials) so a corrupted DB row is
	// distinguishable from a wrong password in logs.
	hasher := bcrypt.New(xbcrypt.MinCost)

	err := hasher.Compare("not-a-bcrypt-hash", "any password")

	if err == nil {
		t.Fatalf("expected error for malformed hash, got nil")
	}
	if errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Errorf("malformed hash must NOT be wrapped as ErrInvalidCredentials, got %v", err)
	}
}
