package invitation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestListInvitationsByMatchUseCase_Execute_ReturnsInvitationsForMatch(t *testing.T) {
	t.Parallel()
	repo := newFakeInvitationRepository()
	ctx := context.Background()
	createdAt := time.Now()
	expiresAt := createdAt.Add(5 * 24 * time.Hour)

	for _, data := range []struct{ id, matchID string }{
		{"inv-1", "match-1"},
		{"inv-2", "match-1"},
		{"inv-3", "match-2"},
	} {
		inv, _ := entities.NewInvitation(
			entities.InvitationID(data.id),
			entities.MatchID(data.matchID),
			"hash-"+data.id, expiresAt, nil, createdAt,
		)
		_ = repo.Save(ctx, inv)
	}

	uc := NewListInvitationsByMatchUseCase(repo)
	invitations, err := uc.Execute(ctx, "match-1")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(invitations) != 2 {
		t.Errorf("expected 2 invitations for match-1, got %d", len(invitations))
	}
}

func TestListInvitationsByMatchUseCase_Execute_PropagatesRepoError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db unreachable")
	repo := newFakeInvitationRepository()
	repo.listByMatchErr = repoErr

	invitations, err := NewListInvitationsByMatchUseCase(repo).Execute(context.Background(), "match-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
	if invitations != nil {
		t.Errorf("expected nil invitations on error, got %v", invitations)
	}
}
