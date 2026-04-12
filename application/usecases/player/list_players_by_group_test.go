package player

import (
	"context"
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
