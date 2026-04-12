package player

import (
	"context"
	"errors"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestGetPlayerUseCase_Execute_ReturnsPlayer_WhenExists(t *testing.T) {
	t.Parallel()
	repo := newFakePlayerRepository()
	existing, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1200)
	_ = repo.Save(context.Background(), existing)

	uc := NewGetPlayerUseCase(repo)
	player, err := uc.Execute(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.Ranking() != 1200 {
		t.Errorf("expected 1200, got %d", player.Ranking())
	}
}

func TestGetPlayerUseCase_Execute_ReturnsError_WhenMissing(t *testing.T) {
	t.Parallel()
	uc := NewGetPlayerUseCase(newFakePlayerRepository())
	_, err := uc.Execute(context.Background(), "nonexistent")
	if !errors.Is(err, domainerrors.ErrPlayerNotFound) {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}
