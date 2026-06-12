package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestNewMatchShareLink_ReturnsActiveLink_WhenInputsAreValid(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)

	link, err := NewMatchShareLink("link-1", "match-1", "hash-abc-123", expiresAt, nil, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if link.ID() != "link-1" {
		t.Errorf("expected ID 'link-1', got %q", link.ID())
	}
	if link.MatchID() != "match-1" {
		t.Errorf("expected MatchID 'match-1', got %q", link.MatchID())
	}
	if link.TokenHash() != "hash-abc-123" {
		t.Errorf("expected token hash 'hash-abc-123', got %q", link.TokenHash())
	}
	if !link.ExpiresAt().Equal(expiresAt) {
		t.Errorf("expected expiresAt %v, got %v", expiresAt, link.ExpiresAt())
	}
	if link.RevokedAt() != nil {
		t.Errorf("expected nil RevokedAt for a fresh link, got %v", link.RevokedAt())
	}
	if !link.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, link.CreatedAt())
	}
	if !link.IsActive(createdAt.Add(time.Hour)) {
		t.Errorf("expected a fresh link to be active before expiry")
	}
}

func TestNewMatchShareLink_AcceptsRevokedAt_WhenRehydrating(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	revokedAt := createdAt.Add(2 * time.Hour)

	link, err := NewMatchShareLink("link-1", "match-1", "hash", expiresAt, &revokedAt, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if link.RevokedAt() == nil || !link.RevokedAt().Equal(revokedAt) {
		t.Errorf("expected RevokedAt %v, got %v", revokedAt, link.RevokedAt())
	}
}

func TestNewMatchShareLink_ReturnsError_WhenInputsAreInvalid(t *testing.T) {
	t.Parallel()

	validCreatedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	validExpiresAt := validCreatedAt.Add(24 * time.Hour)

	cases := []struct {
		name      string
		id        MatchShareLinkID
		matchID   MatchID
		tokenHash string
		expiresAt time.Time
		createdAt time.Time
		wantErr   error
	}{
		{name: "empty id", id: "", matchID: "m-1", tokenHash: "h", expiresAt: validExpiresAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "empty match id", id: "l-1", matchID: "", tokenHash: "h", expiresAt: validExpiresAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "empty token hash", id: "l-1", matchID: "m-1", tokenHash: "", expiresAt: validExpiresAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidID},
		{name: "zero createdAt", id: "l-1", matchID: "m-1", tokenHash: "h", expiresAt: validExpiresAt, createdAt: time.Time{}, wantErr: domainerrors.ErrInvalidDate},
		{name: "zero expiresAt", id: "l-1", matchID: "m-1", tokenHash: "h", expiresAt: time.Time{}, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
		{name: "expiresAt before createdAt", id: "l-1", matchID: "m-1", tokenHash: "h", expiresAt: validCreatedAt.Add(-time.Hour), createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
		{name: "expiresAt equals createdAt", id: "l-1", matchID: "m-1", tokenHash: "h", expiresAt: validCreatedAt, createdAt: validCreatedAt, wantErr: domainerrors.ErrInvalidDate},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			link, err := NewMatchShareLink(testCase.id, testCase.matchID, testCase.tokenHash, testCase.expiresAt, nil, testCase.createdAt)

			if link != nil {
				t.Errorf("expected nil link, got %+v", link)
			}
			if !errors.Is(err, testCase.wantErr) {
				t.Errorf("expected error %v, got %v", testCase.wantErr, err)
			}
		})
	}
}

func TestMatchShareLink_IsActive(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	revokedAt := createdAt.Add(time.Hour)

	cases := []struct {
		name      string
		revokedAt *time.Time
		now       time.Time
		want      bool
	}{
		{name: "before expiration", revokedAt: nil, now: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC), want: true},
		{name: "exactly at expiration", revokedAt: nil, now: expiresAt, want: false},
		{name: "after expiration", revokedAt: nil, now: time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC), want: false},
		{name: "revoked before expiration", revokedAt: &revokedAt, now: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC), want: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			link, _ := NewMatchShareLink("link-1", "match-1", "hash", expiresAt, testCase.revokedAt, createdAt)

			got := link.IsActive(testCase.now)

			if got != testCase.want {
				t.Errorf("expected IsActive=%v, got %v", testCase.want, got)
			}
		})
	}
}

func TestMatchShareLink_Revoke_SetsRevokedAt_WhenActive(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	revokeAt := createdAt.Add(2 * time.Hour)
	link, _ := NewMatchShareLink("link-1", "match-1", "hash", expiresAt, nil, createdAt)

	err := link.Revoke(revokeAt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if link.RevokedAt() == nil || !link.RevokedAt().Equal(revokeAt) {
		t.Errorf("expected RevokedAt %v, got %v", revokeAt, link.RevokedAt())
	}
	if link.IsActive(revokeAt.Add(time.Minute)) {
		t.Errorf("expected a revoked link to be inactive")
	}
}

func TestMatchShareLink_Revoke_ReturnsErrShareLinkInactive_WhenAlreadyRevoked(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(7 * 24 * time.Hour)
	revokedAt := createdAt.Add(time.Hour)
	link, _ := NewMatchShareLink("link-1", "match-1", "hash", expiresAt, &revokedAt, createdAt)

	err := link.Revoke(createdAt.Add(2 * time.Hour))

	if !errors.Is(err, domainerrors.ErrShareLinkInactive) {
		t.Errorf("expected ErrShareLinkInactive, got %v", err)
	}
	if !link.RevokedAt().Equal(revokedAt) {
		t.Errorf("expected original RevokedAt %v to be preserved, got %v", revokedAt, link.RevokedAt())
	}
}

func TestMatchShareLink_Revoke_ReturnsErrShareLinkInactive_WhenExpired(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	link, _ := NewMatchShareLink("link-1", "match-1", "hash", expiresAt, nil, createdAt)

	err := link.Revoke(createdAt.Add(2 * time.Hour))

	if !errors.Is(err, domainerrors.ErrShareLinkInactive) {
		t.Errorf("expected ErrShareLinkInactive, got %v", err)
	}
	if link.RevokedAt() != nil {
		t.Errorf("expected RevokedAt to stay nil on refused revoke, got %v", link.RevokedAt())
	}
}
