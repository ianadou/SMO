package group

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func TestGetGroupUseCase_Execute_ReturnsGroup_WhenItExists(t *testing.T) {
	t.Parallel()

	repo := newFakeGroupRepository()
	existingGroup, _ := entities.NewGroup("group-1", "Existing", "org-1", "", time.Now())
	_ = repo.Save(context.Background(), existingGroup)

	useCase := NewGetGroupUseCase(repo)

	group, err := useCase.Execute(context.Background(), "group-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if group.ID() != "group-1" {
		t.Errorf("expected ID 'group-1', got %q", group.ID())
	}
	if group.Name() != "Existing" {
		t.Errorf("expected name 'Existing', got %q", group.Name())
	}
}

func TestGetGroupUseCase_Execute_ReturnsError_WhenGroupDoesNotExist(t *testing.T) {
	t.Parallel()

	repo := newFakeGroupRepository()
	useCase := NewGetGroupUseCase(repo)

	group, err := useCase.Execute(context.Background(), "nonexistent")

	if group != nil {
		t.Errorf("expected nil group, got %+v", group)
	}
	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}
