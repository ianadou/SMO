package dto

import (
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestInvitationResponseFromEntity_OmitsPlainToken(t *testing.T) {
	t.Parallel()
	createdAt := time.Now()
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, nil, createdAt)

	resp := InvitationResponseFromEntity(inv)
	if resp.ID != "inv-1" {
		t.Errorf("expected 'inv-1', got %q", resp.ID)
	}
	if resp.UsedAt != nil {
		t.Errorf("expected nil UsedAt, got %v", resp.UsedAt)
	}
}

func TestParticipantResponsesFromEntities_MapsAllFields(t *testing.T) {
	t.Parallel()
	confirmedAt := time.Date(2026, 6, 1, 18, 30, 0, 0, time.UTC)
	in := []entities.MatchParticipant{
		{PlayerID: "p-1", PlayerName: "Alice", ConfirmedAt: confirmedAt},
		{PlayerID: "p-2", PlayerName: "Bob", ConfirmedAt: confirmedAt.Add(time.Minute)},
	}

	out := ParticipantResponsesFromEntities(in)

	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	if out[0].PlayerID != "p-1" || out[0].PlayerName != "Alice" || !out[0].ConfirmedAt.Equal(confirmedAt) {
		t.Errorf("entry 0 mismatch: %+v", out[0])
	}
	if out[1].PlayerID != "p-2" || out[1].PlayerName != "Bob" {
		t.Errorf("entry 1 mismatch: %+v", out[1])
	}
}

func TestParticipantResponsesFromEntities_EmptyInput_ReturnsEmptyNotNilSlice(t *testing.T) {
	t.Parallel()
	out := ParticipantResponsesFromEntities(nil)
	if out == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(out) != 0 {
		t.Errorf("expected length 0, got %d", len(out))
	}
}
