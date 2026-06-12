package dto

import (
	"testing"
	"time"

	"github.com/ianadou/smo/application/usecases/invitation"
	"github.com/ianadou/smo/domain/entities"
)

func TestInvitationResponseFromEntity_OmitsPlainToken(t *testing.T) {
	t.Parallel()
	createdAt := time.Now()
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, entities.InvitationResponsePending, nil, nil, createdAt)

	resp := InvitationResponseFromEntity(inv)
	if resp.ID != "inv-1" {
		t.Errorf("expected 'inv-1', got %q", resp.ID)
	}
	if resp.Response != "pending" {
		t.Errorf("expected response 'pending', got %q", resp.Response)
	}
	if resp.RespondedAt != nil {
		t.Errorf("expected nil RespondedAt, got %v", resp.RespondedAt)
	}
}

func TestRespondInvitationResponseFromEntity_ExposesResponseAndTimestamp(t *testing.T) {
	t.Parallel()
	createdAt := time.Now()
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	respondedAt := createdAt.Add(time.Hour)
	inv, _ := entities.NewInvitation("inv-1", "match-1", "p-1", "hash", expiresAt, entities.InvitationResponseYes, &respondedAt, nil, createdAt)

	resp := RespondInvitationResponseFromEntity(inv)

	if resp.Response != "yes" {
		t.Errorf("expected response 'yes', got %q", resp.Response)
	}
	if resp.RespondedAt == nil || !resp.RespondedAt.Equal(respondedAt) {
		t.Errorf("expected RespondedAt %v, got %v", respondedAt, resp.RespondedAt)
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

func TestInvitationContextResponseFromContext_DerivesPresentationFields(t *testing.T) {
	t.Parallel()
	scheduled := time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC)
	expires := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	ctx := &invitation.PageContext{
		OrganizerName:   "Eddin",
		GroupName:       "Les Bras Cassés",
		MatchTitle:      "Foot du jeudi",
		Venue:           "Gerland",
		ScheduledAt:     scheduled,
		MaxParticipants: 10,
		ConfirmedNames:  []string{"Jean Dupont", "Alice"},
		Response:        entities.InvitationResponsePending,
		ExpiresAt:       expires,
		Locked:          false,
		Expired:         false,
	}

	resp := InvitationContextResponseFromContext(ctx)

	if resp.Capacity != "10 (5v5)" {
		t.Errorf("Capacity = %q, want %q", resp.Capacity, "10 (5v5)")
	}
	if resp.ConfirmedCount != 2 {
		t.Errorf("ConfirmedCount = %d, want 2", resp.ConfirmedCount)
	}
	if resp.MaxParticipants != 10 {
		t.Errorf("MaxParticipants = %d, want 10", resp.MaxParticipants)
	}
	if resp.State != "respondable" {
		t.Errorf("State = %q, want respondable", resp.State)
	}
	if resp.Response != "pending" {
		t.Errorf("Response = %q, want pending", resp.Response)
	}
	if !resp.ScheduledAt.Equal(scheduled) {
		t.Errorf("ScheduledAt = %v, want %v", resp.ScheduledAt, scheduled)
	}
	want := []string{"JD", "A"}
	if len(resp.ConfirmedInitials) != len(want) {
		t.Fatalf("ConfirmedInitials = %v, want %v", resp.ConfirmedInitials, want)
	}
	for i := range want {
		if resp.ConfirmedInitials[i] != want[i] {
			t.Errorf("ConfirmedInitials[%d] = %q, want %q", i, resp.ConfirmedInitials[i], want[i])
		}
	}
}

func TestInvitationContextResponseFromContext_ExpiredBeatsLocked(t *testing.T) {
	t.Parallel()
	ctx := &invitation.PageContext{
		MaxParticipants: 10,
		Response:        entities.InvitationResponseYes,
		Expired:         true,
		Locked:          true,
	}

	resp := InvitationContextResponseFromContext(ctx)

	if resp.State != "expired" {
		t.Errorf("State = %q, want expired (expired must take precedence over locked)", resp.State)
	}
}

func TestInvitationContextResponseFromContext_LockedWhenNotExpired(t *testing.T) {
	t.Parallel()
	ctx := &invitation.PageContext{
		MaxParticipants: 10,
		Response:        entities.InvitationResponseYes,
		Expired:         false,
		Locked:          true,
	}

	resp := InvitationContextResponseFromContext(ctx)

	if resp.State != "locked" {
		t.Errorf("State = %q, want locked", resp.State)
	}
}

func TestInvitationContextResponseFromContext_EmptyConfirmed_ReturnsEmptyNotNilSlice(t *testing.T) {
	t.Parallel()
	ctx := &invitation.PageContext{
		MaxParticipants: 10,
		Response:        entities.InvitationResponsePending,
		ConfirmedNames:  nil,
	}

	resp := InvitationContextResponseFromContext(ctx)

	if resp.ConfirmedInitials == nil {
		t.Error("ConfirmedInitials = nil, want empty slice for JSON [] not null")
	}
	if len(resp.ConfirmedInitials) != 0 {
		t.Errorf("ConfirmedInitials len = %d, want 0", len(resp.ConfirmedInitials))
	}
}
