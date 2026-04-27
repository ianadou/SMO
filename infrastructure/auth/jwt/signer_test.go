package jwt_test

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/auth/jwt"
)

const testSecret = "test-secret-test-secret-test-secret-test-secret"

func TestSigner_SignAndVerify_RoundTrip(t *testing.T) {
	t.Parallel()
	signer := jwt.New(testSecret, time.Hour)

	token, err := signer.Sign("org-42")
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	got, err := signer.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if got != "org-42" {
		t.Errorf("expected subject 'org-42', got %q", got)
	}
}

func TestSigner_Verify_RejectsTokenSignedWithDifferentSecret(t *testing.T) {
	t.Parallel()
	good := jwt.New(testSecret, time.Hour)
	bad := jwt.New("a-completely-different-secret", time.Hour)

	token, _ := bad.Sign("org-42")
	_, err := good.Verify(token)

	if !errors.Is(err, domainerrors.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestSigner_Verify_RejectsExpiredToken(t *testing.T) {
	t.Parallel()
	// Lifetime of 1 nanosecond → token is already expired by the time
	// Verify runs.
	signer := jwt.New(testSecret, time.Nanosecond)

	token, _ := signer.Sign("org-42")
	time.Sleep(2 * time.Nanosecond)
	_, err := signer.Verify(token)

	if !errors.Is(err, domainerrors.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestSigner_Verify_RejectsMalformedToken(t *testing.T) {
	t.Parallel()
	signer := jwt.New(testSecret, time.Hour)

	_, err := signer.Verify("definitely-not-a-jwt")

	if !errors.Is(err, domainerrors.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
