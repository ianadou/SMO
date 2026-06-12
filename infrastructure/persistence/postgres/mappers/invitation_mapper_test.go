package mappers

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

func TestInvitationToDomain_ReturnsPendingEntity_WhenRespondedAtIsNull(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	row := generated.Invitations{
		ID:          "inv-1",
		MatchID:     "match-1",
		PlayerID:    "p-1",
		TokenHash:   "abc123",
		ExpiresAt:   pgtype.Timestamptz{Time: createdAt.Add(5 * 24 * time.Hour), Valid: true},
		Response:    "pending",
		RespondedAt: pgtype.Timestamptz{Valid: false},
		CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
	inv, err := InvitationToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Response() != entities.InvitationResponsePending {
		t.Errorf("expected pending, got %q", inv.Response())
	}
	if inv.RespondedAt() != nil {
		t.Errorf("expected nil RespondedAt, got %v", inv.RespondedAt())
	}
}

func TestInvitationToDomain_ReturnsConfirmedEntity_WhenResponseIsYes(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	respondedAt := createdAt.Add(2 * time.Hour)
	row := generated.Invitations{
		ID:          "inv-1",
		MatchID:     "match-1",
		PlayerID:    "p-1",
		TokenHash:   "abc",
		ExpiresAt:   pgtype.Timestamptz{Time: createdAt.Add(5 * 24 * time.Hour), Valid: true},
		Response:    "yes",
		RespondedAt: pgtype.Timestamptz{Time: respondedAt, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
	inv, err := InvitationToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inv.IsConfirmed() {
		t.Errorf("expected confirmed invitation for response=yes")
	}
	if inv.RespondedAt() == nil || !inv.RespondedAt().Equal(respondedAt) {
		t.Errorf("expected RespondedAt %v, got %v", respondedAt, inv.RespondedAt())
	}
}

func TestInvitationToDomain_ReturnsClaimedEntity_WhenClaimedAtIsSet(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	claimedAt := createdAt.Add(3 * time.Hour)
	row := generated.Invitations{
		ID:          "inv-1",
		MatchID:     "match-1",
		PlayerID:    "p-1",
		TokenHash:   "rotated-hash",
		ExpiresAt:   pgtype.Timestamptz{Time: createdAt.Add(5 * 24 * time.Hour), Valid: true},
		Response:    "pending",
		RespondedAt: pgtype.Timestamptz{Valid: false},
		ClaimedAt:   pgtype.Timestamptz{Time: claimedAt, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
	}
	inv, err := InvitationToDomain(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ClaimedAt() == nil || !inv.ClaimedAt().Equal(claimedAt) {
		t.Errorf("expected ClaimedAt %v, got %v", claimedAt, inv.ClaimedAt())
	}
}

func TestInvitationToCreateParams_BuildsParams(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, entities.InvitationResponsePending, nil, nil, createdAt)

	params := InvitationToCreateParams(inv)
	if params.ID != "inv-1" {
		t.Errorf("expected 'inv-1', got %q", params.ID)
	}
	if params.MatchID != "match-1" || params.PlayerID != "p-1" {
		t.Errorf("expected match/player ids mapped, got %+v", params)
	}
}

func TestInvitationToUpdateResponseParams_BuildsParamsWithResponse(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	respondedAt := createdAt.Add(2 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, entities.InvitationResponseYes, &respondedAt, nil, createdAt)

	params := InvitationToUpdateResponseParams(inv)
	if params.Response != "yes" {
		t.Errorf("expected response 'yes', got %q", params.Response)
	}
	if !params.RespondedAt.Valid || !params.RespondedAt.Time.Equal(respondedAt) {
		t.Errorf("expected RespondedAt %v, got %+v", respondedAt, params.RespondedAt)
	}
}

func TestInvitationToClaimParams_BuildsParamsWithRotatedToken(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	claimedAt := createdAt.Add(3 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "p-1", "original-hash", createdAt.Add(5*24*time.Hour), entities.InvitationResponsePending, nil, nil, createdAt)
	if err := inv.Claim("rotated-hash", claimedAt); err != nil {
		t.Fatalf("Claim: %v", err)
	}

	params := InvitationToClaimParams(inv)
	if params.ID != "inv-1" {
		t.Errorf("expected 'inv-1', got %q", params.ID)
	}
	if params.TokenHash != "rotated-hash" {
		t.Errorf("expected rotated token hash, got %q", params.TokenHash)
	}
	if !params.ClaimedAt.Valid || !params.ClaimedAt.Time.Equal(claimedAt) {
		t.Errorf("expected ClaimedAt %v, got %+v", claimedAt, params.ClaimedAt)
	}
}
