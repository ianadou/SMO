package vote

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestGetVoteUseCase_Execute_ReturnsVote_WhenExists(t *testing.T) {
	t.Parallel()
	repo := newFakeVoteRepository()
	v, _ := entities.NewVote("v-1", "m-1", "p-1", "p-2", 4, time.Now())
	_ = repo.Save(context.Background(), v)

	uc := NewGetVoteUseCase(repo)
	found, err := uc.Execute(context.Background(), "v-1")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if found.Score() != 4 {
		t.Errorf("expected 4, got %d", found.Score())
	}
}

func TestGetVoteUseCase_Execute_ReturnsErrVoteNotFound(t *testing.T) {
	t.Parallel()
	uc := NewGetVoteUseCase(newFakeVoteRepository())
	_, err := uc.Execute(context.Background(), "missing")
	if !errors.Is(err, domainerrors.ErrVoteNotFound) {
		t.Errorf("expected ErrVoteNotFound, got %v", err)
	}
}
