package mappers

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

func TestInvitationToDomain_ReturnsEntity_WhenUsedAtIsNull(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	row := generated.Invitations{
		ID:        "inv-1",
		MatchID:   "match-1",
		TokenHash: "abc123",
		ExpiresAt: pgtype.Timestamptz{Time: createdAt.Add(5 * 24 * time.Hour), Valid: true},
		UsedAt:    pgtype.Timestamptz{Valid: false},
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
	inv, err := InvitationToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.UsedAt() != nil {
		t.Errorf("expected nil UsedAt, got %v", inv.UsedAt())
	}
}

func TestInvitationToDomain_ReturnsEntity_WhenUsedAtIsSet(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	usedAt := createdAt.Add(2 * time.Hour)
	row := generated.Invitations{
		ID:        "inv-1",
		MatchID:   "match-1",
		TokenHash: "abc",
		ExpiresAt: pgtype.Timestamptz{Time: createdAt.Add(5 * 24 * time.Hour), Valid: true},
		UsedAt:    pgtype.Timestamptz{Time: usedAt, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
	inv, err := InvitationToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.UsedAt() == nil || !inv.UsedAt().Equal(usedAt) {
		t.Errorf("expected UsedAt %v, got %v", usedAt, inv.UsedAt())
	}
}

func TestInvitationToCreateParams_BuildsParams(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "hash", expiresAt, nil, createdAt)

	params := InvitationToCreateParams(inv)
	if params.ID != "inv-1" {
		t.Errorf("expected 'inv-1', got %q", params.ID)
	}
	if params.UsedAt.Valid {
		t.Errorf("expected UsedAt to be NULL, got valid=true")
	}
}

func TestInvitationToMarkAsUsedParams_BuildsParamsWithUsedAt(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	usedAt := createdAt.Add(2 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "hash", expiresAt, &usedAt, createdAt)

	params := InvitationToMarkAsUsedParams(inv)
	if !params.UsedAt.Valid || !params.UsedAt.Time.Equal(usedAt) {
		t.Errorf("expected UsedAt %v, got %+v", usedAt, params.UsedAt)
	}
}
