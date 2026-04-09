package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewInvitation_ReturnsInvitation_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)

	inv, err := NewInvitation("inv-1", "match-1", "hash-abc-123", expiresAt, nil, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if inv.ID() != "inv-1" {
		t.Errorf("expected ID 'inv-1', got %q", inv.ID())
	}
	if inv.MatchID() != "match-1" {
		t.Errorf("expected MatchID 'match-1', got %q", inv.MatchID())
	}
	if inv.TokenHash() != "hash-abc-123" {
		t.Errorf("expected token hash, got %q", inv.TokenHash())
	}
	if !inv.ExpiresAt().Equal(expiresAt) {
		t.Errorf("expected expiresAt %v, got %v", expiresAt, inv.ExpiresAt())
	}
	if inv.UsedAt() != nil {
		t.Errorf("expected nil UsedAt for new invitation, got %v", inv.UsedAt())
	}
	if inv.IsUsed() {
		t.Errorf("expected new invitation to not be used")
	}
}

func TestNewInvitation_AcceptsUsedAtWhenProvided(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	usedAt := createdAt.Add(2 * time.Hour)

	inv, err := NewInvitation("inv-1", "match-1", "hash", expiresAt, &usedAt, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !inv.IsUsed() {
		t.Errorf("expected invitation to be marked as used")
	}
	if inv.UsedAt() == nil || !inv.UsedAt().Equal(usedAt) {
		t.Errorf("expected UsedAt %v, got %v", usedAt, inv.UsedAt())
	}
}

func TestInvitation_IsExpired(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	inv, _ := NewInvitation("inv-1", "match-1", "hash", expiresAt, nil, createdAt)

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "before expiration", now: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC), want: false},
		{name: "exactly at expiration", now: expiresAt, want: true},
		{name: "after expiration", now: time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC), want: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := inv.IsExpired(testCase.now)

			if got != testCase.want {
				t.Errorf("expected IsExpired=%v, got %v", testCase.want, got)
			}
		})
	}
}

func TestNewInvitation_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	validCreatedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	validExpiresAt := validCreatedAt.Add(24 * time.Hour)

	cases := []struct {
		name      string
		id        InvitationID
		matchID   MatchID
		tokenHash string
		expiresAt time.Time
		createdAt time.Time
		wantErr   error
	}{
		{name: "empty id", id: "", matchID: "m-1", tokenHash: "h", expiresAt: validExpiresAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "empty match id", id: "i-1", matchID: "", tokenHash: "h", expiresAt: validExpiresAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "empty token hash", id: "i-1", matchID: "m-1", tokenHash: "", expiresAt: validExpiresAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "zero createdAt", id: "i-1", matchID: "m-1", tokenHash: "h", expiresAt: validExpiresAt, createdAt: time.Time{}, wantErr: domainerrors.ErrInvalidDate},
		{name: "zero expiresAt", id: "i-1", matchID: "m-1", tokenHash: "h", expiresAt: time.Time{}, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
		{name: "expiresAt before createdAt", id: "i-1", matchID: "m-1", tokenHash: "h", expiresAt: validCreatedAt.Add(-time.Hour), createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
		{name: "expiresAt equals createdAt", id: "i-1", matchID: "m-1", tokenHash: "h", expiresAt: validCreatedAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			inv, err := NewInvitation(testCase.id, testCase.matchID, testCase.tokenHash, testCase.expiresAt, nil, testCase.createdAt)

			if inv != nil {
				t.Errorf("expected nil invitation, got %+v", inv)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}
