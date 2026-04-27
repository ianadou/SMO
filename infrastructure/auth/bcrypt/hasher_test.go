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
