package player

import (
	"context"
	"errors"
	"testing"

	"github.com/ianadou/smo/domain/entities"
)

func TestListPlayersByGroupUseCase_Execute_ReturnsPlayersInGroup(t *testing.T) {
	t.Parallel()
	repo := newFakePlayerRepository()
	ctx := context.Background()
	for _, data := range []struct{ id, name string }{
		{"p-1", "Alice"}, {"p-2", "Bob"}, {"p-3", "Carol"},
	} {
		p, _ := entities.NewPlayer(entities.PlayerID(data.id), "g-1", data.name, 1000)
		_ = repo.Save(ctx, p)
	}

	uc := NewListPlayersByGroupUseCase(repo)
	players, err := uc.Execute(ctx, "g-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 3 {
		t.Errorf("expected 3 players, got %d", len(players))
	}
}

func TestListPlayersByGroupUseCase_Execute_PropagatesRepoError(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("db unreachable")
	repo := newFakePlayerRepository()
	repo.listByGroupErr = repoErr

	players, err := NewListPlayersByGroupUseCase(repo).Execute(context.Background(), "g-1")

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped repo error, got %v", err)
	}
	if players != nil {
		t.Errorf("expected nil players on error, got %v", players)
	}
}
