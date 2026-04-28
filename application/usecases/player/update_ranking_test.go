package player

import (
	"context"
	"errors"
	"testing"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestUpdatePlayerRankingUseCase_Execute_PersistsNewRanking(t *testing.T) {
	t.Parallel()
	repo := newFakePlayerRepository()
	existing, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1000)
	_ = repo.Save(context.Background(), existing)

	uc := NewUpdatePlayerRankingUseCase(repo)
	updated, err := uc.Execute(context.Background(), "p-1", 1300)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Ranking() != 1300 {
		t.Errorf("expected ranking 1300, got %d", updated.Ranking())
	}

	found, _ := repo.FindByID(context.Background(), "p-1")
	if found.Ranking() != 1300 {
		t.Errorf("expected persisted ranking 1300, got %d", found.Ranking())
	}
}

func TestUpdatePlayerRankingUseCase_Execute_ReturnsError_WhenMissing(t *testing.T) {
	t.Parallel()
	uc := NewUpdatePlayerRankingUseCase(newFakePlayerRepository())
	_, err := uc.Execute(context.Background(), "nonexistent", 1300)
	if !errors.Is(err, domainerrors.ErrPlayerNotFound) {
		t.Errorf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestUpdatePlayerRankingUseCase_Execute_ReturnsError_WhenPersistFails(t *testing.T) {
	t.Parallel()
	repoErr := errors.New("disk full")
	repo := newFakePlayerRepository()
	existing, _ := entities.NewPlayer("p-1", "g-1", "Alice", 1000)
	_ = repo.Save(context.Background(), existing)
	repo.updateRankingErr = repoErr

	_, err := NewUpdatePlayerRankingUseCase(repo).Execute(context.Background(), "p-1", 1500)

	if !errors.Is(err, repoErr) {
		t.Errorf("expected wrapped persist error, got %v", err)
	}
}
