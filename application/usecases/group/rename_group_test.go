package group

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

func seedGroup(t *testing.T, repo *fakeGroupRepository, id entities.GroupID, owner entities.OrganizerID) {
	t.Helper()
	seeded, err := entities.NewGroup(id, "Foot du jeudi", owner, "",
		time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("seed group: %v", err)
	}
	if saveErr := repo.Save(context.Background(), seeded); saveErr != nil {
		t.Fatalf("seed save: %v", saveErr)
	}
}

func TestRenameGroupUseCase_RenamesAndPersists_WhenOwnerRequests(t *testing.T) {
	t.Parallel()
	repo := newFakeGroupRepository()
	seedGroup(t, repo, "g-1", "org-1")
	uc := NewRenameGroupUseCase(repo)

	renamed, err := uc.Execute(context.Background(), "g-1", "org-1", "Foot du vendredi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if renamed.Name() != "Foot du vendredi" {
		t.Errorf("expected renamed group, got %q", renamed.Name())
	}
	stored, _ := repo.FindByID(context.Background(), "g-1")
	if stored.Name() != "Foot du vendredi" {
		t.Errorf("expected persisted name, got %q", stored.Name())
	}
}

func TestRenameGroupUseCase_ReturnsErrGroupNotFound_WhenNotOwner(t *testing.T) {
	t.Parallel()
	repo := newFakeGroupRepository()
	seedGroup(t, repo, "g-1", "org-1")
	uc := NewRenameGroupUseCase(repo)

	_, err := uc.Execute(context.Background(), "g-1", "org-intruder", "Pris !")

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound for a foreign group, got %v", err)
	}
	stored, _ := repo.FindByID(context.Background(), "g-1")
	if stored.Name() != "Foot du jeudi" {
		t.Errorf("foreign rename must not persist, got %q", stored.Name())
	}
}

func TestRenameGroupUseCase_ReturnsErrGroupNotFound_WhenGroupMissing(t *testing.T) {
	t.Parallel()
	uc := NewRenameGroupUseCase(newFakeGroupRepository())

	_, err := uc.Execute(context.Background(), "ghost", "org-1", "Peu importe")

	if !errors.Is(err, domainerrors.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestRenameGroupUseCase_ReturnsErrInvalidName_WhenNameInvalid(t *testing.T) {
	t.Parallel()
	repo := newFakeGroupRepository()
	seedGroup(t, repo, "g-1", "org-1")
	uc := NewRenameGroupUseCase(repo)

	_, err := uc.Execute(context.Background(), "g-1", "org-1", "   ")

	if !errors.Is(err, domainerrors.ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}
