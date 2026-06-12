package dto

import (
	"testing"
	"time"

	"github.com/ianadou/smo/application/usecases/sharelink"
	"github.com/ianadou/smo/domain/entities"
)

func TestMatchShareLinkResponseFromResult_ExposesPlainTokenAndExpiry(t *testing.T) {
	t.Parallel()
	createdAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(5 * 24 * time.Hour)
	link, _ := entities.NewMatchShareLink("link-1", "match-1", "hash", expiresAt, nil, createdAt)

	resp := MatchShareLinkResponseFromResult(&sharelink.GenerateMatchShareLinkResult{
		ShareLink:  link,
		PlainToken: "plain-token-once",
	})

	if resp.Token != "plain-token-once" {
		t.Errorf("expected plain token 'plain-token-once', got %q", resp.Token)
	}
	if !resp.ExpiresAt.Equal(expiresAt) {
		t.Errorf("expected expiry %v, got %v", expiresAt, resp.ExpiresAt)
	}
}

func TestShareLinkContextResponseFromContext_MapsMatchAndRosterFields(t *testing.T) {
	t.Parallel()
	scheduledAt := time.Date(2026, 6, 20, 19, 30, 0, 0, time.UTC)

	resp := ShareLinkContextResponseFromContext(&sharelink.PageContext{
		MatchID:         "match-1",
		OrganizerName:   "Eddin",
		GroupName:       "Les Bras Cassés",
		MatchTitle:      "Match du vendredi",
		Venue:           "Five Bezons",
		ScheduledAt:     scheduledAt,
		MatchStatus:     entities.MatchStatusOpen,
		MaxParticipants: 10,
		ConfirmedNames:  []string{"Karim Benzema", "Zinedine"},
		Roster: []sharelink.RosterEntry{
			{PlayerID: "p-1", PlayerName: "Karim Benzema", State: sharelink.RosterStateResponded},
			{PlayerID: "p-2", PlayerName: "Sofiane", State: sharelink.RosterStateClaimed},
			{PlayerID: "p-3", PlayerName: "Yacine", State: sharelink.RosterStateClaimable},
		},
	})

	if resp.MatchID != "match-1" {
		t.Errorf("expected match id 'match-1', got %q", resp.MatchID)
	}
	if resp.OrganizerName != "Eddin" {
		t.Errorf("expected organizer 'Eddin', got %q", resp.OrganizerName)
	}
	if resp.GroupName != "Les Bras Cassés" {
		t.Errorf("expected group 'Les Bras Cassés', got %q", resp.GroupName)
	}
	if resp.MatchStatus != "open" {
		t.Errorf("expected status 'open', got %q", resp.MatchStatus)
	}
	if resp.Capacity != "10 (5v5)" {
		t.Errorf("expected capacity '10 (5v5)', got %q", resp.Capacity)
	}
	if resp.ConfirmedCount != 2 {
		t.Errorf("expected 2 confirmed, got %d", resp.ConfirmedCount)
	}
	if len(resp.ConfirmedInitials) != 2 || resp.ConfirmedInitials[0] != "KB" || resp.ConfirmedInitials[1] != "Z" {
		t.Errorf("expected initials [KB Z], got %v", resp.ConfirmedInitials)
	}
	if len(resp.Roster) != 3 {
		t.Fatalf("expected 3 roster entries, got %d", len(resp.Roster))
	}
	if resp.Roster[0].PlayerID != "p-1" || resp.Roster[0].PlayerName != "Karim Benzema" || resp.Roster[0].State != "responded" {
		t.Errorf("unexpected first roster entry: %+v", resp.Roster[0])
	}
	if resp.Roster[1].State != "claimed" || resp.Roster[2].State != "claimable" {
		t.Errorf("expected states claimed/claimable, got %q/%q", resp.Roster[1].State, resp.Roster[2].State)
	}
}

func TestShareLinkContextResponseFromContext_EmptyRosterAndConfirmed_ReturnsEmptyNotNilSlices(t *testing.T) {
	t.Parallel()

	resp := ShareLinkContextResponseFromContext(&sharelink.PageContext{
		MatchID:         "match-1",
		MatchStatus:     entities.MatchStatusOpen,
		MaxParticipants: 10,
	})

	if resp.ConfirmedInitials == nil || len(resp.ConfirmedInitials) != 0 {
		t.Errorf("expected empty non-nil initials, got %v", resp.ConfirmedInitials)
	}
	if resp.Roster == nil || len(resp.Roster) != 0 {
		t.Errorf("expected empty non-nil roster, got %v", resp.Roster)
	}
}
