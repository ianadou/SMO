package match

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func newSetTeamsUseCase(t *testing.T) (
	*SetTeamsUseCase,
	*fakeMatchRepository,
	*fakeInvitationRepoForGenerate,
) {
	t.Helper()
	matchRepo := newFakeMatchRepository()
	invRepo := newFakeInvitationRepoForGenerate()
	uc := NewSetTeamsUseCase(matchRepo, invRepo)
	return uc, matchRepo, invRepo
}

func seedConfirmedParticipants(
	invRepo *fakeInvitationRepoForGenerate,
	ids ...entities.PlayerID,
) {
	for _, id := range ids {
		invRepo.participants[generateTestMatchID] = append(
			invRepo.participants[generateTestMatchID],
			entities.MatchParticipant{
				PlayerID:    id,
				PlayerName:  string(id),
				ConfirmedAt: time.Now(),
			},
		)
	}
}

func TestSetTeams_PersistsExactPartition(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo := newSetTeamsUseCase(t)
	ctx := context.Background()

	seedOpenMatch(t, matchRepo)
	seedConfirmedParticipants(invRepo, "p1", "p2", "p3", "p4")

	result, err := uc.Execute(ctx, generateTestMatchID,
		[]entities.PlayerID{"p1", "p2"}, []entities.PlayerID{"p3", "p4"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !elementsMatch(result.TeamA(), []entities.PlayerID{"p1", "p2"}) {
		t.Errorf("expected TeamA {p1,p2}, got %v", result.TeamA())
	}
	if !elementsMatch(result.TeamB(), []entities.PlayerID{"p3", "p4"}) {
		t.Errorf("expected TeamB {p3,p4}, got %v", result.TeamB())
	}
	if matchRepo.replaceTeamsCalls != 1 {
		t.Fatalf("expected ReplaceTeams called once, got %d", matchRepo.replaceTeamsCalls)
	}
}

func TestSetTeams_Rejected_WhenNotAPartitionOfConfirmed(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo := newSetTeamsUseCase(t)
	ctx := context.Background()

	seedOpenMatch(t, matchRepo)
	seedConfirmedParticipants(invRepo, "p1", "p2")

	_, err := uc.Execute(ctx, generateTestMatchID,
		[]entities.PlayerID{"p1"}, []entities.PlayerID{"pX"})

	if !errors.Is(err, domainerrors.ErrInvalidAssignment) {
		t.Errorf("expected ErrInvalidAssignment, got %v", err)
	}
}

func TestSetTeams_Rejected_WhenImbalanced(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo := newSetTeamsUseCase(t)
	ctx := context.Background()

	seedOpenMatch(t, matchRepo)
	seedConfirmedParticipants(invRepo, "p1", "p2", "p3", "p4")

	_, err := uc.Execute(ctx, generateTestMatchID,
		[]entities.PlayerID{"p1", "p2", "p3"}, []entities.PlayerID{"p4"})

	if !errors.Is(err, domainerrors.ErrInvalidAssignment) {
		t.Errorf("expected ErrInvalidAssignment, got %v", err)
	}
}

func TestSetTeams_Rejected_WhenMatchNotOpen(t *testing.T) {
	t.Parallel()
	uc, matchRepo, invRepo := newSetTeamsUseCase(t)
	ctx := context.Background()

	m, err := entities.NewMatch(generateTestMatchID, "g-1", "Test", "Venue",
		time.Now().Add(time.Hour), time.Now())
	if err != nil {
		t.Fatalf("seed: NewMatch: %v", err)
	}
	if saveErr := matchRepo.Save(ctx, m); saveErr != nil {
		t.Fatalf("seed: Save: %v", saveErr)
	}
	seedConfirmedParticipants(invRepo, "p1", "p2", "p3", "p4")

	_, err = uc.Execute(ctx, generateTestMatchID,
		[]entities.PlayerID{"p1", "p2"}, []entities.PlayerID{"p3", "p4"})

	if !errors.Is(err, domainerrors.ErrTeamsNotEditable) {
		t.Errorf("expected ErrTeamsNotEditable, got %v", err)
	}
}

func elementsMatch(got, want []entities.PlayerID) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[entities.PlayerID]int, len(got))
	for _, id := range got {
		seen[id]++
	}
	for _, id := range want {
		if seen[id] == 0 {
			return false
		}
		seen[id]--
	}
	return true
}
