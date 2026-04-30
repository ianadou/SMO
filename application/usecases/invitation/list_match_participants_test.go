package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func seedConfirmedInvitation(t *testing.T, repo *fakeInvitationRepository, id, matchID, playerID string, usedAt time.Time) {
	t.Helper()
	createdAt := usedAt.Add(-time.Hour)
	expiresAt := createdAt.Add(2 * 24 * time.Hour)
	inv, err := entities.NewInvitation(
		entities.InvitationID(id),
		entities.MatchID(matchID),
		entities.PlayerID(playerID),
		"hash-"+id,
		expiresAt,
		&usedAt,
		createdAt,
	)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = repo.Save(context.Background(), inv)
}

func TestListMatchParticipantsUseCase_Execute_ReturnsConfirmedOnly(t *testing.T) {
	t.Parallel()
	repo := newFakeInvitationRepository()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	seedConfirmedInvitation(t, repo, "inv-1", "match-1", "alice", now)
	seedConfirmedInvitation(t, repo, "inv-2", "match-1", "bob", now.Add(time.Minute))

	// A pending invitation (not used) for the same match — must be excluded.
	createdAt := now.Add(-time.Hour)
	pending, _ := entities.NewInvitation(
		"inv-pending", "match-1", "ghost", "hash-pending",
		now.Add(24*time.Hour), nil, createdAt,
	)
	_ = repo.Save(context.Background(), pending)

	uc := NewListMatchParticipantsUseCase(repo)
	participants, err := uc.Execute(context.Background(), "match-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(participants) != 2 {
		t.Errorf("expected 2 confirmed participants (pending excluded), got %d", len(participants))
	}
}

func TestListMatchParticipantsUseCase_Execute_ScopesToTheGivenMatch(t *testing.T) {
	t.Parallel()
	repo := newFakeInvitationRepository()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	seedConfirmedInvitation(t, repo, "inv-A", "match-A", "alice", now)
	seedConfirmedInvitation(t, repo, "inv-B", "match-B", "bob", now)

	uc := NewListMatchParticipantsUseCase(repo)
	participantsA, _ := uc.Execute(context.Background(), "match-A")
	participantsB, _ := uc.Execute(context.Background(), "match-B")

	if len(participantsA) != 1 || participantsA[0].PlayerID != "alice" {
		t.Errorf("match-A should have only alice, got %+v", participantsA)
	}
	if len(participantsB) != 1 || participantsB[0].PlayerID != "bob" {
		t.Errorf("match-B should have only bob, got %+v", participantsB)
	}
}

func TestListMatchParticipantsUseCase_Execute_ReturnsEmptySlice_WhenNoConfirmedYet(t *testing.T) {
	t.Parallel()
	repo := newFakeInvitationRepository()

	uc := NewListMatchParticipantsUseCase(repo)
	participants, err := uc.Execute(context.Background(), "match-empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if participants == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(participants) != 0 {
		t.Errorf("expected 0 participants, got %d", len(participants))
	}
}

func TestListMatchParticipantsUseCase_Execute_RejectsEmptyMatchID(t *testing.T) {
	t.Parallel()
	uc := NewListMatchParticipantsUseCase(newFakeInvitationRepository())
	_, err := uc.Execute(context.Background(), "")
	if err == nil {
		t.Error("expected error on empty match id")
	}
}

func TestListMatchParticipantsUseCase_Execute_PropagatesRepoError(t *testing.T) {
	t.Parallel()
	repo := newFakeInvitationRepository()
	repo.listConfirmedErr = errors.New("db down")
	uc := NewListMatchParticipantsUseCase(repo)
	_, err := uc.Execute(context.Background(), "match-1")
	if err == nil || !errors.Is(err, repo.listConfirmedErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
}
