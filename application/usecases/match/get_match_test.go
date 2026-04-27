package match

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestGetMatchUseCase_Execute_ReturnsMatch_WhenItExists(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	existing, _ := entities.NewMatch(
		"match-1", "group-1", "Existing", "Venue",
		time.Now().Add(24*time.Hour), time.Now(),
	)
	_ = repo.Save(context.Background(), existing)

	useCase := NewGetMatchUseCase(repo)

	match, err := useCase.Execute(context.Background(), "match-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if match.Title() != "Existing" {
		t.Errorf("expected title 'Existing', got %q", match.Title())
	}
}

func TestGetMatchUseCase_Execute_ReturnsError_WhenMatchDoesNotExist(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	useCase := NewGetMatchUseCase(repo)

	match, err := useCase.Execute(context.Background(), "nonexistent")

	if match != nil {
		t.Errorf("expected nil match, got %+v", match)
	}
	if !errors.Is(err, domainerrors.ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}
