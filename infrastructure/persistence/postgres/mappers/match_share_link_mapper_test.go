package mappers

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

func TestMatchShareLinkToDomain_ReturnsActiveLink_WhenRevokedAtIsNull(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	row := generated.MatchShareLinks{
		ID:        "link-1",
		MatchID:   "match-1",
		TokenHash: "abc123",
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		RevokedAt: pgtype.Timestamptz{Valid: false},
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
	link, err := MatchShareLinkToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.RevokedAt() != nil {
		t.Errorf("expected nil RevokedAt, got %v", link.RevokedAt())
	}
	if !link.IsActive(createdAt.Add(time.Hour)) {
		t.Error("expected an unrevoked, unexpired link to be active")
	}
}

func TestMatchShareLinkToDomain_ReturnsRevokedLink_WhenRevokedAtIsSet(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	revokedAt := createdAt.Add(2 * time.Hour)
	row := generated.MatchShareLinks{
		ID:        "link-1",
		MatchID:   "match-1",
		TokenHash: "abc123",
		ExpiresAt: pgtype.Timestamptz{Time: createdAt.Add(5 * 24 * time.Hour), Valid: true},
		RevokedAt: pgtype.Timestamptz{Time: revokedAt, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
	link, err := MatchShareLinkToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.RevokedAt() == nil || !link.RevokedAt().Equal(revokedAt) {
		t.Errorf("expected RevokedAt %v, got %v", revokedAt, link.RevokedAt())
	}
	if link.IsActive(revokedAt.Add(time.Minute)) {
		t.Error("expected a revoked link to be inactive")
	}
}

func TestMatchShareLinkToCreateParams_BuildsParams(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	link, _ := entities.NewMatchShareLink("link-1", "match-1", "hash", expiresAt, nil, createdAt)

	params := MatchShareLinkToCreateParams(link)
	if params.ID != "link-1" {
		t.Errorf("expected 'link-1', got %q", params.ID)
	}
	if params.MatchID != "match-1" || params.TokenHash != "hash" {
		t.Errorf("expected match id and hash mapped, got %+v", params)
	}
	if !params.ExpiresAt.Valid || !params.ExpiresAt.Time.Equal(expiresAt) {
		t.Errorf("expected ExpiresAt %v, got %+v", expiresAt, params.ExpiresAt)
	}
}

func TestMatchShareLinkToUpdateParams_BuildsParamsWithRevocation(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	link, _ := entities.NewMatchShareLink("link-1", "match-1", "hash", createdAt.Add(48*time.Hour), nil, createdAt)
	revokedAt := createdAt.Add(2 * time.Hour)
	if err := link.Revoke(revokedAt); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	params := MatchShareLinkToUpdateParams(link)
	if params.ID != "link-1" {
		t.Errorf("expected 'link-1', got %q", params.ID)
	}
	if !params.RevokedAt.Valid || !params.RevokedAt.Time.Equal(revokedAt) {
		t.Errorf("expected RevokedAt %v, got %+v", revokedAt, params.RevokedAt)
	}
}
