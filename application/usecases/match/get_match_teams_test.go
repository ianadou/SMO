package match

import (
	"context"
	"testing"

	"github.com/ianadou/smo/domain/entities"
)

func TestGetMatchTeams_ReturnsBothTeamsWithNames(t *testing.T) {
	repo := newFakeMatchRepository()
	repo.teamMembers = map[entities.MatchID][]entities.MatchTeamMember{
		"m1": {
			{PlayerID: "p1", PlayerName: "Alex L.", Team: "A", Slot: 0},
			{PlayerID: "p2", PlayerName: "Inès R.", Team: "B", Slot: 0},
		},
	}
	uc := NewGetMatchTeamsUseCase(repo)

	got, err := uc.Execute(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 members, got %d", len(got))
	}
	if got[0].PlayerName != "Alex L." || got[0].Team != "A" {
		t.Errorf("unexpected first member: %+v", got[0])
	}
}
