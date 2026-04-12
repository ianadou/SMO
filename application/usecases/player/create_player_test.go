package player

import (
	"context"
	"errors"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestCreatePlayerUseCase_Execute_CreatesPlayerWithDefaultRanking(t *testing.T) {
	t.Parallel()
	repo := newFakePlayerRepository()
	idGen := newFakeIDGenerator("p-1")
	uc := NewCreatePlayerUseCase(repo, idGen)

	player, err := uc.Execute(context.Background(), CreatePlayerInput{GroupID: "g-1", Name: "Alice"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if player.Ranking() != entities.DefaultPlayerRanking() {
		t.Errorf("expected default ranking %d, got %d", entities.DefaultPlayerRanking(), player.Ranking())
	}
	if player.Name() != "Alice" {
		t.Errorf("expected 'Alice', got %q", player.Name())
	}
}

func TestCreatePlayerUseCase_Execute_ReturnsError_WhenNameIsEmpty(t *testing.T) {
	t.Parallel()
	repo := newFakePlayerRepository()
	uc := NewCreatePlayerUseCase(repo, newFakeIDGenerator("p-1"))

	_, err := uc.Execute(context.Background(), CreatePlayerInput{GroupID: "g-1", Name: ""})

	if !errors.Is(err, domainerrors.ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}
