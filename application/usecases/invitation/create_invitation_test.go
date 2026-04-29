package invitation

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCreateInvitationUseCase_Execute_ReturnsPlainTokenAndInvitation(t *testing.T) {
	t.Parallel()
	repo := newFakeInvitationRepository()
	tokens := newFakeTokenService("plain-token-abc")
	idGen := newFakeIDGenerator("inv-1")
	clock := newFakeClock(time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))
	uc := NewCreateInvitationUseCase(repo, tokens, idGen, clock)

	result, err := uc.Execute(context.Background(), CreateInvitationInput{MatchID: "match-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlainToken != "plain-token-abc" {
		t.Errorf("expected plain token 'plain-token-abc', got %q", result.PlainToken)
	}
	// Stored hash must equal SHA-256 of the plain token.
	if result.Invitation.TokenHash() != tokens.HashToken("plain-token-abc") {
		t.Errorf("stored hash does not match hash of plain token")
	}
}

func TestCreateInvitationUseCase_Execute_AppliesDefaultValidityWindow_WhenExpiresAtIsZero(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	uc := NewCreateInvitationUseCase(
		newFakeInvitationRepository(),
		newFakeTokenService("tok"),
		newFakeIDGenerator("inv-1"),
		newFakeClock(now),
	)

	result, err := uc.Execute(context.Background(), CreateInvitationInput{MatchID: "match-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := now.Add(DefaultInvitationValidityDuration)
	if !result.Invitation.ExpiresAt().Equal(expected) {
		t.Errorf("expected expiresAt %v, got %v", expected, result.Invitation.ExpiresAt())
	}
}

func TestCreateInvitationUseCase_Execute_UsesProvidedExpiresAt(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	customExpires := now.Add(24 * time.Hour)
	uc := NewCreateInvitationUseCase(
		newFakeInvitationRepository(),
		newFakeTokenService("tok"),
		newFakeIDGenerator("inv-1"),
		newFakeClock(now),
	)

	result, err := uc.Execute(context.Background(), CreateInvitationInput{
		MatchID:   "match-1",
		ExpiresAt: customExpires,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Invitation.ExpiresAt().Equal(customExpires) {
		t.Errorf("expected expiresAt %v, got %v", customExpires, result.Invitation.ExpiresAt())
	}
}

func TestCreateInvitationUseCase_Execute_PropagatesSaveError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	repo := newFakeInvitationRepository()
	repo.saveErr = repoErr
	uc := NewCreateInvitationUseCase(
		repo,
		newFakeTokenService("tok"),
		newFakeIDGenerator("inv-1"),
		newFakeClock(time.Now()),
	)

	_, err := uc.Execute(context.Background(), CreateInvitationInput{MatchID: "match-1"})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}
